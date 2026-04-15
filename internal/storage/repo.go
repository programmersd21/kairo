package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/util"
)

type Repository struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Repository, error) {
	if strings.TrimSpace(path) == "" {
		stateDir, err := util.AppStateDir("kairo")
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(stateDir, 0o755); err != nil {
			return nil, err
		}
		path = filepath.Join(stateDir, "kairo.db")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	// Pragmas are applied via DSN. modernc.org/sqlite uses "sqlite" driver.
	dsn := fmt.Sprintf("file:%s?cache=shared&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", filepath.ToSlash(path))

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error { return r.db.Close() }

func (r *Repository) CreateTask(ctx context.Context, t core.Task) (core.Task, error) {
	now := time.Now()
	if t.ID == "" {
		id, err := core.NewID(now)
		if err != nil {
			return core.Task{}, err
		}
		t.ID = id
	}
	t.Status = core.Status(strings.ToLower(string(t.Status)))
	if t.Status == "" {
		t.Status = core.StatusTodo
	}
	t.Priority = t.Priority.Clamp()
	t.Tags = t.NormalizedTags()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	if err := t.Validate(); err != nil {
		return core.Task{}, err
	}

	return t, withTx(ctx, r.db, func(tx *sql.Tx) error {
		var deadline any
		if t.Deadline != nil {
			deadline = t.Deadline.UTC().UnixMilli()
		} else {
			deadline = nil
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO tasks (id, title, description, priority, deadline_ms, status, created_at_ms, updated_at_ms)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, t.ID, t.Title, t.Description, int(t.Priority), deadline, string(t.Status), t.CreatedAt.UTC().UnixMilli(), t.UpdatedAt.UTC().UnixMilli())
		if err != nil {
			return err
		}

		return setTaskTags(ctx, tx, t.ID, t.Tags)
	})
}

func (r *Repository) UpdateTask(ctx context.Context, id string, patch core.TaskPatch) (core.Task, error) {
	existing, err := r.GetTask(ctx, id)
	if err != nil {
		return core.Task{}, err
	}
	updated := patch.ApplyTo(existing)
	updated.Status = core.Status(strings.ToLower(string(updated.Status)))
	updated.Priority = updated.Priority.Clamp()
	updated.Tags = updated.NormalizedTags()
	updated.UpdatedAt = time.Now()
	if err := updated.Validate(); err != nil {
		return core.Task{}, err
	}

	return updated, withTx(ctx, r.db, func(tx *sql.Tx) error {
		var deadline any
		if updated.Deadline != nil {
			deadline = updated.Deadline.UTC().UnixMilli()
		} else {
			deadline = nil
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE tasks
			SET title=?, description=?, priority=?, deadline_ms=?, status=?, updated_at_ms=?
			WHERE id=? AND deleted_at_ms IS NULL
		`, updated.Title, updated.Description, int(updated.Priority), deadline, string(updated.Status), updated.UpdatedAt.UTC().UnixMilli(), updated.ID)
		if err != nil {
			return err
		}
		return setTaskTags(ctx, tx, updated.ID, updated.Tags)
	})
}

func (r *Repository) DeleteTask(ctx context.Context, id string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		now := time.Now().UTC().UnixMilli()
		_, err := tx.ExecContext(ctx, `UPDATE tasks SET deleted_at_ms=?, updated_at_ms=? WHERE id=? AND deleted_at_ms IS NULL`, now, now, id)
		return err
	})
}

type Tombstone struct {
	ID        string
	DeletedAt time.Time
	UpdatedAt time.Time
}

func (r *Repository) SyncSnapshot(ctx context.Context) ([]core.Task, []Tombstone, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, priority, deadline_ms, status, created_at_ms, updated_at_ms
		FROM tasks
		WHERE deleted_at_ms IS NULL
		ORDER BY updated_at_ms DESC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var tasks []core.Task
	for rows.Next() {
		var (
			id         string
			title      string
			desc       string
			priority   int
			deadlineMs sql.NullInt64
			status     string
			createdMs  int64
			updatedMs  int64
		)
		if err := rows.Scan(&id, &title, &desc, &priority, &deadlineMs, &status, &createdMs, &updatedMs); err != nil {
			return nil, nil, err
		}
		var deadline *time.Time
		if deadlineMs.Valid {
			d := time.UnixMilli(deadlineMs.Int64).UTC()
			deadline = &d
		}
		tasks = append(tasks, core.Task{
			ID:          id,
			Title:       title,
			Description: desc,
			Priority:    core.Priority(priority).Clamp(),
			Deadline:    deadline,
			Status:      core.Status(status),
			CreatedAt:   time.UnixMilli(createdMs).UTC(),
			UpdatedAt:   time.UnixMilli(updatedMs).UTC(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// tags in bulk
	idSet := make([]any, 0, len(tasks))
	for _, t := range tasks {
		idSet = append(idSet, t.ID)
	}
	if len(idSet) > 0 {
		holders := strings.TrimRight(strings.Repeat("?,", len(idSet)), ",")
		tagRows, err := r.db.QueryContext(ctx, `
			SELECT tt.task_id, g.name
			FROM task_tags tt
			JOIN tags g ON g.id=tt.tag_id
			WHERE tt.task_id IN (`+holders+`)
			ORDER BY g.name ASC
		`, idSet...)
		if err == nil {
			defer tagRows.Close()
			tagsByID := map[string][]string{}
			for tagRows.Next() {
				var taskID, name string
				if err := tagRows.Scan(&taskID, &name); err != nil {
					return nil, nil, err
				}
				tagsByID[taskID] = append(tagsByID[taskID], name)
			}
			for i := range tasks {
				tasks[i].Tags = tagsByID[tasks[i].ID]
			}
		}
	}

	tRows, err := r.db.QueryContext(ctx, `
		SELECT id, deleted_at_ms, updated_at_ms
		FROM tasks
		WHERE deleted_at_ms IS NOT NULL
	`)
	if err != nil {
		return tasks, nil, nil
	}
	defer tRows.Close()
	var tomb []Tombstone
	for tRows.Next() {
		var id string
		var delMs int64
		var updMs int64
		if err := tRows.Scan(&id, &delMs, &updMs); err != nil {
			return nil, nil, err
		}
		tomb = append(tomb, Tombstone{
			ID:        id,
			DeletedAt: time.UnixMilli(delMs).UTC(),
			UpdatedAt: time.UnixMilli(updMs).UTC(),
		})
	}
	return tasks, tomb, tRows.Err()
}

func (r *Repository) ApplyTombstone(ctx context.Context, t Tombstone) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		var existingUpdated sql.NullInt64
		_ = tx.QueryRowContext(ctx, `SELECT updated_at_ms FROM tasks WHERE id=?`, t.ID).Scan(&existingUpdated)
		if existingUpdated.Valid && existingUpdated.Int64 > t.UpdatedAt.UTC().UnixMilli() {
			return nil
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tasks (id, title, description, priority, deadline_ms, status, created_at_ms, updated_at_ms, deleted_at_ms)
			VALUES (?, '', '', 0, NULL, 'todo', ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				deleted_at_ms=excluded.deleted_at_ms,
				updated_at_ms=excluded.updated_at_ms
		`, t.ID, t.UpdatedAt.UTC().UnixMilli(), t.UpdatedAt.UTC().UnixMilli(), t.DeletedAt.UTC().UnixMilli())
		return err
	})
}

func (r *Repository) GetTask(ctx context.Context, id string) (core.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, priority, deadline_ms, status, created_at_ms, updated_at_ms
		FROM tasks
		WHERE id=? AND deleted_at_ms IS NULL
	`, id)
	t, err := scanTask(row)
	if err != nil {
		return core.Task{}, err
	}
	t.Tags, _ = r.taskTags(ctx, id)
	return t, nil
}

type ListOptions struct {
	Filter core.Filter
	Limit  int
}

func (r *Repository) ListTasks(ctx context.Context, opt ListOptions) ([]core.Task, error) {
	where := []string{"t.deleted_at_ms IS NULL"}
	args := []any{}
	if len(opt.Filter.Statuses) > 0 {
		holders := make([]string, 0, len(opt.Filter.Statuses))
		for _, st := range opt.Filter.Statuses {
			holders = append(holders, "?")
			args = append(args, string(st))
		}
		where = append(where, "t.status IN ("+strings.Join(holders, ",")+")")
	}
	if opt.Filter.From != nil {
		where = append(where, "t.deadline_ms IS NOT NULL AND t.deadline_ms >= ?")
		args = append(args, opt.Filter.From.UTC().UnixMilli())
	}
	if opt.Filter.To != nil {
		where = append(where, "t.deadline_ms IS NOT NULL AND t.deadline_ms < ?")
		args = append(args, opt.Filter.To.UTC().UnixMilli())
	}
	if opt.Filter.IncludeNilDeadline && opt.Filter.From == nil && opt.Filter.To == nil {
		// no deadline constraints; include nil by default.
	}
	if opt.Filter.Tag != "" {
		where = append(where, "EXISTS (SELECT 1 FROM task_tags tt JOIN tags g ON g.id=tt.tag_id WHERE tt.task_id=t.id AND g.name=?)")
		args = append(args, core.NormalizeTag(opt.Filter.Tag))
	}
	if opt.Filter.Priority != nil {
		where = append(where, "t.priority=?")
		args = append(args, int((*opt.Filter.Priority).Clamp()))
	}

	limit := opt.Limit
	if limit <= 0 {
		limit = 500
	}
	query := `
		SELECT t.id, t.title, t.description, t.priority, t.deadline_ms, t.status, t.created_at_ms, t.updated_at_ms
		FROM tasks t
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY
			CASE WHEN t.deadline_ms IS NULL THEN 1 ELSE 0 END,
			t.deadline_ms ASC,
			t.updated_at_ms DESC
		LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []core.Task{}
	for rows.Next() {
		var (
			id         string
			title      string
			desc       string
			priority   int
			deadlineMs sql.NullInt64
			status     string
			createdMs  int64
			updatedMs  int64
		)
		if err := rows.Scan(&id, &title, &desc, &priority, &deadlineMs, &status, &createdMs, &updatedMs); err != nil {
			return nil, err
		}
		var deadline *time.Time
		if deadlineMs.Valid {
			d := time.UnixMilli(deadlineMs.Int64).UTC()
			deadline = &d
		}
		out = append(out, core.Task{
			ID:          id,
			Title:       title,
			Description: desc,
			Priority:    core.Priority(priority).Clamp(),
			Deadline:    deadline,
			Status:      core.Status(status),
			CreatedAt:   time.UnixMilli(createdMs).UTC(),
			UpdatedAt:   time.UnixMilli(updatedMs).UTC(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load tags in bulk.
	idSet := make([]any, 0, len(out))
	for _, t := range out {
		idSet = append(idSet, t.ID)
	}
	if len(idSet) == 0 {
		return out, nil
	}
	holders := strings.TrimRight(strings.Repeat("?,", len(idSet)), ",")
	tagRows, err := r.db.QueryContext(ctx, `
		SELECT tt.task_id, g.name
		FROM task_tags tt
		JOIN tags g ON g.id=tt.tag_id
		WHERE tt.task_id IN (`+holders+`)
		ORDER BY g.name ASC
	`, idSet...)
	if err != nil {
		return out, nil
	}
	defer tagRows.Close()

	tagsByID := map[string][]string{}
	for tagRows.Next() {
		var taskID, name string
		if err := tagRows.Scan(&taskID, &name); err != nil {
			return nil, err
		}
		tagsByID[taskID] = append(tagsByID[taskID], name)
	}
	for i := range out {
		out[i].Tags = tagsByID[out[i].ID]
	}
	return out, nil
}

func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM tags ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *Repository) AllTasks(ctx context.Context) ([]core.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, priority, deadline_ms, status, created_at_ms, updated_at_ms
		FROM tasks
		WHERE deleted_at_ms IS NULL
		ORDER BY updated_at_ms DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.Task
	for rows.Next() {
		var (
			id         string
			title      string
			desc       string
			priority   int
			deadlineMs sql.NullInt64
			status     string
			createdMs  int64
			updatedMs  int64
		)
		if err := rows.Scan(&id, &title, &desc, &priority, &deadlineMs, &status, &createdMs, &updatedMs); err != nil {
			return nil, err
		}
		var deadline *time.Time
		if deadlineMs.Valid {
			d := time.UnixMilli(deadlineMs.Int64).UTC()
			deadline = &d
		}
		out = append(out, core.Task{
			ID:          id,
			Title:       title,
			Description: desc,
			Priority:    core.Priority(priority).Clamp(),
			Deadline:    deadline,
			Status:      core.Status(status),
			CreatedAt:   time.UnixMilli(createdMs).UTC(),
			UpdatedAt:   time.UnixMilli(updatedMs).UTC(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// tags in bulk
	idSet := make([]any, 0, len(out))
	for _, t := range out {
		idSet = append(idSet, t.ID)
	}
	if len(idSet) == 0 {
		return out, nil
	}
	holders := strings.TrimRight(strings.Repeat("?,", len(idSet)), ",")
	tagRows, err := r.db.QueryContext(ctx, `
		SELECT tt.task_id, g.name
		FROM task_tags tt
		JOIN tags g ON g.id=tt.tag_id
		WHERE tt.task_id IN (`+holders+`)
		ORDER BY g.name ASC
	`, idSet...)
	if err != nil {
		return out, nil
	}
	defer tagRows.Close()
	tagsByID := map[string][]string{}
	for tagRows.Next() {
		var taskID, name string
		if err := tagRows.Scan(&taskID, &name); err != nil {
			return nil, err
		}
		tagsByID[taskID] = append(tagsByID[taskID], name)
	}
	for i := range out {
		out[i].Tags = tagsByID[out[i].ID]
	}
	return out, nil
}

// UpsertTask inserts or updates a task by ID. The update is applied only if
// incoming UpdatedAt is newer than the stored record (last-writer-wins).
func (r *Repository) UpsertTask(ctx context.Context, t core.Task) error {
	now := time.Now()
	if t.ID == "" {
		id, err := core.NewID(now)
		if err != nil {
			return err
		}
		t.ID = id
	}
	t.Status = core.Status(strings.ToLower(string(t.Status)))
	if t.Status == "" {
		t.Status = core.StatusTodo
	}
	t.Priority = t.Priority.Clamp()
	t.Tags = t.NormalizedTags()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}
	if err := t.Validate(); err != nil {
		return err
	}

	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		var existingUpdated sql.NullInt64
		_ = tx.QueryRowContext(ctx, `SELECT updated_at_ms FROM tasks WHERE id=? AND deleted_at_ms IS NULL`, t.ID).Scan(&existingUpdated)
		if existingUpdated.Valid && existingUpdated.Int64 > t.UpdatedAt.UTC().UnixMilli() {
			return nil
		}
		var deadline any
		if t.Deadline != nil {
			deadline = t.Deadline.UTC().UnixMilli()
		} else {
			deadline = nil
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tasks (id, title, description, priority, deadline_ms, status, created_at_ms, updated_at_ms)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				title=excluded.title,
				description=excluded.description,
				priority=excluded.priority,
				deadline_ms=excluded.deadline_ms,
				status=excluded.status,
				created_at_ms=excluded.created_at_ms,
				updated_at_ms=excluded.updated_at_ms,
				deleted_at_ms=NULL
		`, t.ID, t.Title, t.Description, int(t.Priority), deadline, string(t.Status), t.CreatedAt.UTC().UnixMilli(), t.UpdatedAt.UTC().UnixMilli())
		if err != nil {
			return err
		}
		return setTaskTags(ctx, tx, t.ID, t.Tags)
	})
}

func (r *Repository) taskTags(ctx context.Context, id string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.name
		FROM task_tags tt
		JOIN tags g ON g.id=tt.tag_id
		WHERE tt.task_id=?
		ORDER BY g.name ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanTask(s scanner) (core.Task, error) {
	var (
		id         string
		title      string
		desc       string
		priority   int
		deadlineMs sql.NullInt64
		status     string
		createdMs  int64
		updatedMs  int64
	)
	if err := s.Scan(&id, &title, &desc, &priority, &deadlineMs, &status, &createdMs, &updatedMs); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.Task{}, err
		}
		return core.Task{}, err
	}
	var deadline *time.Time
	if deadlineMs.Valid {
		d := time.UnixMilli(deadlineMs.Int64).UTC()
		deadline = &d
	}
	return core.Task{
		ID:          id,
		Title:       title,
		Description: desc,
		Priority:    core.Priority(priority).Clamp(),
		Deadline:    deadline,
		Status:      core.Status(status),
		CreatedAt:   time.UnixMilli(createdMs).UTC(),
		UpdatedAt:   time.UnixMilli(updatedMs).UTC(),
	}, nil
}

func withTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func setTaskTags(ctx context.Context, tx *sql.Tx, taskID string, tags []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM task_tags WHERE task_id=?`, taskID); err != nil {
		return err
	}
	for _, t := range tags {
		if t == "" {
			continue
		}
		var tagID int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO UPDATE SET name=excluded.name RETURNING id`, t).Scan(&tagID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?)`, taskID, tagID); err != nil {
			return err
		}
	}
	return nil
}
