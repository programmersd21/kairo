// Package service provides the unified task service layer - the single source of truth
// for all task operations (TUI, Lua, CLI, Automation).
//
// This ensures consistent behavior and validation across all interfaces.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/hooks"
	"github.com/programmersd21/kairo/internal/storage"
)

// TaskService is the core service interface for all task operations.
// It is the single source of truth used by:
// - Bubble Tea UI
// - Lua plugins
// - CLI API
// - Any external automation
type TaskService interface {
	// Create creates a new task with validation and ID generation.
	Create(ctx context.Context, task core.Task) (core.Task, error)

	// GetByID retrieves a single task by ID.
	GetByID(ctx context.Context, id string) (core.Task, error)

	// Update applies a patch to an existing task.
	Update(ctx context.Context, id string, patch core.TaskPatch) (core.Task, error)

	// Delete soft-deletes a task.
	Delete(ctx context.Context, id string) error

	// DeleteAll soft-deletes all active tasks.
	DeleteAll(ctx context.Context) error

	// List retrieves tasks filtered by the given options.
	List(ctx context.Context, filter core.Filter) ([]core.Task, error)

	// ListAll retrieves all non-deleted tasks.
	ListAll(ctx context.Context) ([]core.Task, error)

	// ListTags retrieves all distinct tags.
	ListTags(ctx context.Context) ([]string, error)

	// GetSnapshot returns all active tasks and tombstones for sync.
	GetSnapshot(ctx context.Context) ([]core.Task, []storage.Tombstone, error)

	// ApplyTombstone merges a remote deletion.
	ApplyTombstone(ctx context.Context, tombstone storage.Tombstone) error

	// UpsertTask inserts or replaces a task (for sync merge).
	UpsertTask(ctx context.Context, task core.Task) error

	// Prune performs a hard delete of soft-deleted tasks and optimizes the database.
	Prune(ctx context.Context) error

	// Hooks returns the event manager for this service.
	Hooks() *hooks.Manager

	// Repo returns the underlying repository.
	Repo() *storage.Repository
}

// taskService implements TaskService using a repository backend.
type taskService struct {
	repo  *storage.Repository
	hooks *hooks.Manager
}

// New creates a new task service.
func New(repo *storage.Repository, hks *hooks.Manager) TaskService {
	return &taskService{
		repo:  repo,
		hooks: hks,
	}
}

// Hooks returns the event manager.
func (s *taskService) Hooks() *hooks.Manager {
	return s.hooks
}

// Repo returns the underlying repository.
func (s *taskService) Repo() *storage.Repository {
	return s.repo
}

// Create creates a new task with validation and ID generation.
func (s *taskService) Create(ctx context.Context, task core.Task) (core.Task, error) {
	// Normalize and validate
	if err := task.Validate(); err != nil {
		return core.Task{}, fmt.Errorf("invalid task: %w", err)
	}

	// Let repository handle ID generation, timestamps, and persistence
	created, err := s.repo.CreateTask(ctx, task)
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to create task: %w", err)
	}

	// Emit event
	s.hooks.TaskCreated(created)

	return created, nil
}

// GetByID retrieves a task by ID.
func (s *taskService) GetByID(ctx context.Context, id string) (core.Task, error) {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// Update applies a patch to a task.
func (s *taskService) Update(ctx context.Context, id string, patch core.TaskPatch) (core.Task, error) {
	// Repository handles validation of patch and atomic update
	updated, err := s.repo.UpdateTask(ctx, id, patch)
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to update task: %w", err)
	}

	// Recurrence logic: if marked done and is recurring, create next instance
	if patch.Status != nil && *patch.Status == core.StatusDone && updated.Recurrence != core.RecurrenceNone {
		// Use deadline as reference if available, else now
		ref := time.Now()
		if updated.Deadline != nil {
			ref = *updated.Deadline
		}

		nextDue := updated.NextOccurrence(ref)
		if nextDue != nil {
			nextTask := updated
			nextTask.ID = "" // New ID
			nextTask.Status = core.StatusTodo
			nextTask.Deadline = nextDue
			nextTask.CreatedAt = time.Time{}
			nextTask.UpdatedAt = time.Time{}

			_, _ = s.repo.CreateTask(ctx, nextTask)
		}
	}

	// Emit event
	s.hooks.TaskUpdated(updated, patch)

	return updated, nil
}

// Delete soft-deletes a task.
func (s *taskService) Delete(ctx context.Context, id string) error {
	if err := s.repo.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Emit event
	s.hooks.TaskDeleted(id)

	return nil
}

// DeleteAll soft-deletes all tasks.
func (s *taskService) DeleteAll(ctx context.Context) error {
	if err := s.repo.DeleteAllTasks(ctx); err != nil {
		return fmt.Errorf("failed to delete all tasks: %w", err)
	}

	// Emit event
	s.hooks.TaskDeleteAll()

	return nil
}

// List retrieves filtered tasks.
func (s *taskService) List(ctx context.Context, filter core.Filter) ([]core.Task, error) {
	tasks, err := s.repo.ListTasks(ctx, storage.ListOptions{
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

// ListAll retrieves all tasks.
func (s *taskService) ListAll(ctx context.Context) ([]core.Task, error) {
	tasks, err := s.repo.AllTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all tasks: %w", err)
	}
	return tasks, nil
}

// ListTags retrieves all distinct tags.
func (s *taskService) ListTags(ctx context.Context) ([]string, error) {
	tags, err := s.repo.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return tags, nil
}

// GetSnapshot returns all active tasks and tombstones for sync.
func (s *taskService) GetSnapshot(ctx context.Context) ([]core.Task, []storage.Tombstone, error) {
	tasks, tombstones, err := s.repo.SyncSnapshot(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get snapshot: %w", err)
	}
	return tasks, tombstones, nil
}

// ApplyTombstone merges a remote deletion.
func (s *taskService) ApplyTombstone(ctx context.Context, tombstone storage.Tombstone) error {
	if err := s.repo.ApplyTombstone(ctx, tombstone); err != nil {
		return fmt.Errorf("failed to apply tombstone: %w", err)
	}
	return nil
}

// UpsertTask inserts or replaces a task (for sync merge).
func (s *taskService) UpsertTask(ctx context.Context, task core.Task) error {
	if err := s.repo.UpsertTask(ctx, task); err != nil {
		return fmt.Errorf("failed to upsert task: %w", err)
	}
	return nil
}

// Prune performs a hard delete and vacuum.
func (s *taskService) Prune(ctx context.Context) error {
	if err := s.repo.Prune(ctx); err != nil {
		return fmt.Errorf("failed to prune tasks: %w", err)
	}
	if err := s.repo.Vacuum(ctx); err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}
	return nil
}
