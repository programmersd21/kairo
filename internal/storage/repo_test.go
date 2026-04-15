package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/programmersd21/kairo/internal/core"
)

func TestCreateListUpdateDelete(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "kairo.db")
	r, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	created, err := r.CreateTask(ctx, core.Task{
		Title:       "Ship v1",
		Description: "Do the thing",
		Tags:        []string{"Release", "#Go"},
		Priority:    core.P2,
		Status:      core.StatusTodo,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatalf("expected id")
	}

	ts, err := r.ListTasks(ctx, ListOptions{Filter: core.Filter{Statuses: []core.Status{core.StatusTodo, core.StatusDoing}, IncludeNilDeadline: true}})
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 1 {
		t.Fatalf("expected 1 task, got %d", len(ts))
	}
	if len(ts[0].Tags) != 2 || ts[0].Tags[0] != "go" || ts[0].Tags[1] != "release" {
		t.Fatalf("unexpected tags: %#v", ts[0].Tags)
	}

	newTitle := "Ship v1.0"
	st := core.StatusDoing
	updated, err := r.UpdateTask(ctx, created.ID, core.TaskPatch{Title: &newTitle, Status: &st})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != newTitle || updated.Status != st {
		t.Fatalf("unexpected update: %#v", updated)
	}

	if err := r.DeleteTask(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	ts, err = r.AllTasks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 0 {
		t.Fatalf("expected 0 tasks after delete, got %d", len(ts))
	}
}

func TestUpsertAndTombstone(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "kairo.db")
	r, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	now := time.Now().UTC()
	id, _ := core.NewID(now)
	t1 := core.Task{
		ID:        id,
		Title:     "A",
		Status:    core.StatusTodo,
		Priority:  core.P1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Minute),
	}
	if err := r.UpsertTask(ctx, t1); err != nil {
		t.Fatal(err)
	}

	// Older update should not win.
	tOld := t1
	tOld.Title = "OLD"
	tOld.UpdatedAt = now.Add(-2 * time.Minute)
	if err := r.UpsertTask(ctx, tOld); err != nil {
		t.Fatal(err)
	}
	got, err := r.GetTask(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "A" {
		t.Fatalf("expected last-writer-wins, got %q", got.Title)
	}

	// Tombstone newer should delete.
	tb := Tombstone{ID: id, DeletedAt: now, UpdatedAt: now.Add(10 * time.Second)}
	if err := r.ApplyTombstone(ctx, tb); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetTask(ctx, id); err == nil {
		t.Fatalf("expected deleted task to be missing")
	}
}
