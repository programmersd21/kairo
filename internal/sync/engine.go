package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/storage"
)

type Strategy string

const (
	StrategyOurs   Strategy = "ours"
	StrategyTheirs Strategy = "theirs"
)

type Engine struct {
	repo *storage.Repository

	repoPath string
	remote   string
	branch   string
	strategy Strategy
	autoPush bool
}

func New(repo *storage.Repository, repoPath, remote, branch string, strategy Strategy, autoPush bool) *Engine {
	if branch == "" {
		branch = "main"
	}
	if remote == "" {
		remote = "origin"
	}
	if strategy != StrategyTheirs {
		strategy = StrategyOurs
	}
	return &Engine{
		repo:     repo,
		repoPath: repoPath,
		remote:   remote,
		branch:   branch,
		strategy: strategy,
		autoPush: autoPush,
	}
}

func (e *Engine) Enabled() bool { return strings.TrimSpace(e.repoPath) != "" }

func (e *Engine) SyncNow(ctx context.Context) error {
	if !e.Enabled() {
		return errors.New("sync repo_path not set")
	}
	if err := os.MkdirAll(e.repoPath, 0o755); err != nil {
		return err
	}

	if err := e.ensureGitRepo(ctx); err != nil {
		return err
	}
	_ = e.pull(ctx)

	tasks, tomb, err := e.repo.SyncSnapshot(ctx)
	if err != nil {
		return err
	}
	if err := e.writeSnapshot(tasks, tomb); err != nil {
		return err
	}
	if err := e.gitAdd(ctx); err != nil {
		return err
	}
	changed, err := e.hasChanges(ctx)
	if err != nil {
		return err
	}
	if changed {
		msg := "kairo: sync " + time.Now().Format("2006-01-02 15:04:05")
		if err := e.gitCommit(ctx, msg); err != nil {
			return err
		}
	}
	if e.autoPush {
		_ = e.push(ctx)
	}

	_ = e.applyFromRepo(ctx)
	return nil
}

func (e *Engine) ensureGitRepo(ctx context.Context) error {
	gitDir := filepath.Join(e.repoPath, ".git")
	if st, err := os.Stat(gitDir); err == nil && st.IsDir() {
		return nil
	}
	if err := e.git(ctx, "init"); err != nil {
		return err
	}
	_ = e.git(ctx, "checkout", "-b", e.branch)
	return nil
}

func (e *Engine) pull(ctx context.Context) error {
	// Best-effort: pull if remote exists.
	if err := e.git(ctx, "remote", "get-url", e.remote); err != nil {
		return err
	}
	opt := "-X"
	val := string(e.strategy)
	return e.git(ctx, "pull", "--no-rebase", opt, val, e.remote, e.branch)
}

func (e *Engine) push(ctx context.Context) error {
	if err := e.git(ctx, "remote", "get-url", e.remote); err != nil {
		return err
	}
	return e.git(ctx, "push", e.remote, e.branch)
}

func (e *Engine) gitAdd(ctx context.Context) error {
	return e.git(ctx, "add", "-A")
}

func (e *Engine) hasChanges(ctx context.Context) (bool, error) {
	out, err := e.gitOut(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

func (e *Engine) gitCommit(ctx context.Context, msg string) error {
	// Ensure identity isn't a hard error; git will still commit if configured globally.
	_ = e.git(ctx, "config", "user.name", "kairo")
	_ = e.git(ctx, "config", "user.email", "kairo@local")
	return e.git(ctx, "commit", "-m", msg)
}

func (e *Engine) writeSnapshot(tasks []core.Task, tomb []storage.Tombstone) error {
	taskDir := filepath.Join(e.repoPath, "tasks")
	tombDir := filepath.Join(e.repoPath, "tombstones")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(tombDir, 0o755); err != nil {
		return err
	}

	// Write tasks.
	taskIDs := make(map[string]struct{}, len(tasks))
	for _, t := range tasks {
		taskIDs[t.ID] = struct{}{}
		b, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(taskDir, t.ID+".json"), b); err != nil {
			return err
		}
	}
	// Remove stale task files.
	_ = filepath.WalkDir(taskDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		id := strings.TrimSuffix(d.Name(), ".json")
		if _, ok := taskIDs[id]; !ok {
			_ = os.Remove(path)
		}
		return nil
	})

	// Write tombstones.
	tombIDs := make(map[string]struct{}, len(tomb))
	for _, t := range tomb {
		tombIDs[t.ID] = struct{}{}
		w := map[string]any{
			"id":         t.ID,
			"deleted_at": t.DeletedAt.UTC().Format(time.RFC3339Nano),
			"updated_at": t.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		b, err := json.MarshalIndent(w, "", "  ")
		if err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(tombDir, t.ID+".json"), b); err != nil {
			return err
		}
	}
	_ = filepath.WalkDir(tombDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		id := strings.TrimSuffix(d.Name(), ".json")
		if _, ok := tombIDs[id]; !ok {
			_ = os.Remove(path)
		}
		return nil
	})

	return nil
}

func (e *Engine) applyFromRepo(ctx context.Context) error {
	taskDir := filepath.Join(e.repoPath, "tasks")
	entries, err := os.ReadDir(taskDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(taskDir, ent.Name()))
		if err != nil {
			continue
		}
		var t core.Task
		if err := json.Unmarshal(b, &t); err != nil {
			continue
		}
		_ = e.repo.UpsertTask(ctx, t)
	}

	tombDir := filepath.Join(e.repoPath, "tombstones")
	ents, _ := os.ReadDir(tombDir)
	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(tombDir, ent.Name()))
		if err != nil {
			continue
		}
		var w struct {
			ID        string `json:"id"`
			DeletedAt string `json:"deleted_at"`
			UpdatedAt string `json:"updated_at"`
		}
		if err := json.Unmarshal(b, &w); err != nil {
			continue
		}
		del, err1 := time.Parse(time.RFC3339Nano, w.DeletedAt)
		upd, err2 := time.Parse(time.RFC3339Nano, w.UpdatedAt)
		if err1 != nil || err2 != nil {
			continue
		}
		_ = e.repo.ApplyTombstone(ctx, storage.Tombstone{ID: w.ID, DeletedAt: del, UpdatedAt: upd})
	}
	return nil
}

func atomicWrite(path string, b []byte) error {
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp", filepath.Base(path)))
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (e *Engine) git(ctx context.Context, args ...string) error {
	_, err := e.gitOut(ctx, args...)
	return err
}

func (e *Engine) gitOut(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
