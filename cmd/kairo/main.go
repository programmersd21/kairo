package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/programmersd21/kairo/internal/app"
	"github.com/programmersd21/kairo/internal/config"
	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/core/codec"
	"github.com/programmersd21/kairo/internal/storage"
	ksync "github.com/programmersd21/kairo/internal/sync"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kairo: config:", err)
		os.Exit(2)
	}

	repo, err := storage.Open(ctx, cfg.Storage.Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kairo: storage:", err)
		os.Exit(2)
	}
	defer repo.Close()

	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "export":
			if err := runExport(ctx, repo, os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "kairo export:", err)
				os.Exit(2)
			}
			return
		case "import":
			if err := runImport(ctx, repo, os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "kairo import:", err)
				os.Exit(2)
			}
			return
		case "sync":
			if !cfg.Sync.Enabled || strings.TrimSpace(cfg.Sync.RepoPath) == "" {
				fmt.Fprintln(os.Stderr, "kairo sync: enable sync.repo_path in config")
				os.Exit(2)
			}
			eng := ksync.New(repo, cfg.Sync.RepoPath, cfg.Sync.Remote, cfg.Sync.Branch, ksync.Strategy(cfg.Sync.Strategy), cfg.Sync.AutoPush)
			if err := eng.SyncNow(ctx); err != nil {
				fmt.Fprintln(os.Stderr, "kairo sync:", err)
				os.Exit(2)
			}
			return
		}
	}

	m, err := app.New(ctx, cfg, repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kairo:", err)
		os.Exit(2)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "kairo:", err)
		os.Exit(1)
	}
}

func runExport(ctx context.Context, repo *storage.Repository, args []string) error {
	format := "json"
	out := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = strings.ToLower(args[i+1])
				i++
			}
		case "--out":
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		}
	}
	tasks, err := repo.AllTasks(ctx)
	if err != nil {
		return err
	}

	var b []byte
	switch format {
	case "json":
		b, err = codec.MarshalJSON(tasks)
	case "md", "markdown":
		b = codec.MarshalMarkdown(tasks)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if err != nil {
		return err
	}

	if out == "" {
		_, err = os.Stdout.Write(b)
		if err == nil && len(b) > 0 && b[len(b)-1] != '\n' {
			_, _ = os.Stdout.Write([]byte("\n"))
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, b, 0o644)
}

func runImport(ctx context.Context, repo *storage.Repository, args []string) error {
	format := "json"
	in := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = strings.ToLower(args[i+1])
				i++
			}
		case "--in":
			if i+1 < len(args) {
				in = args[i+1]
				i++
			}
		}
	}
	if in == "" {
		return errors.New("--in required")
	}
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	var tasks []core.Task
	switch format {
	case "json":
		tasks, err = codec.UnmarshalJSON(b)
	case "md", "markdown":
		tasks, err = codec.UnmarshalMarkdown(b)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := repo.UpsertTask(ctx, t); err != nil {
			return err
		}
	}
	return nil
}
