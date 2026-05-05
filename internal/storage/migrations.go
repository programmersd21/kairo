package storage

import (
	"context"
	"database/sql"
)

func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`); err != nil {
		return err
	}
	var v int
	_ = db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&v)

	steps := []struct {
		version int
		sql     string
	}{
		{1, `
			CREATE TABLE IF NOT EXISTS tasks (
				id TEXT PRIMARY KEY,
				title TEXT NOT NULL,
				description TEXT NOT NULL DEFAULT '',
				priority INTEGER NOT NULL DEFAULT 0,
				deadline_ms INTEGER NULL,
				status TEXT NOT NULL DEFAULT 'todo',
				created_at_ms INTEGER NOT NULL,
				updated_at_ms INTEGER NOT NULL,
				deleted_at_ms INTEGER NULL
			);

			CREATE TABLE IF NOT EXISTS tags (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE
			);

			CREATE TABLE IF NOT EXISTS task_tags (
				task_id TEXT NOT NULL,
				tag_id INTEGER NOT NULL,
				PRIMARY KEY(task_id, tag_id),
				FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE,
				FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
			CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline_ms);
			CREATE INDEX IF NOT EXISTS idx_tasks_updated ON tasks(updated_at_ms);
			CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag_id);
			CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id);
			CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
		`},
		{2, `
			ALTER TABLE tasks ADD COLUMN recurrence TEXT NOT NULL DEFAULT 'none';
			ALTER TABLE tasks ADD COLUMN recurrence_weekly TEXT NULL;
			ALTER TABLE tasks ADD COLUMN recurrence_monthly INTEGER NOT NULL DEFAULT 0;
		`},
		{3, `
			ALTER TABLE tasks ADD COLUMN parent_id TEXT;
			ALTER TABLE tasks ADD COLUMN collapsed INTEGER NOT NULL DEFAULT 0;
			CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
		`},
	}

	for _, s := range steps {
		if v >= s.version {
			continue
		}
		if _, err := db.ExecContext(ctx, s.sql); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (?)`, s.version); err != nil {
			return err
		}
		v = s.version
	}
	return nil
}
