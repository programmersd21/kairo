package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/hooks"
	"github.com/programmersd21/kairo/internal/storage"
)

func TestParentProjectInheritance(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "kairo.db")
	repo, err := storage.Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := New(repo, hooks.New())

	parent, err := svc.Create(ctx, core.Task{
		Title:      "Parent",
		Project:    "Work",
		Status:     core.StatusTodo,
		Recurrence: core.RecurrenceNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	child, err := svc.Create(ctx, core.Task{
		Title:      "Child (should inherit project)",
		ParentID:   parent.ID,
		Status:     core.StatusTodo,
		Recurrence: core.RecurrenceNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if child.Project != "Work" {
		t.Fatalf("expected child to inherit project 'Work', got %q", child.Project)
	}

	childExplicit, err := svc.Create(ctx, core.Task{
		Title:      "Child (explicit project)",
		ParentID:   parent.ID,
		Project:    "Personal",
		Status:     core.StatusTodo,
		Recurrence: core.RecurrenceNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if childExplicit.Project != "Personal" {
		t.Fatalf("expected child project 'Personal', got %q", childExplicit.Project)
	}
}
