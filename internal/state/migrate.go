package state

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

const schemaVersion = `
	CREATE TABLE IF NOT EXISTS schema_version (
		migration  TEXT PRIMARY KEY,
		sha256     BLOB NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (datetime('now', 'utc'))
	);
`

//go:embed migrations/*.sql
var migrations embed.FS

func (s *Store) Migrate(ctx context.Context) error {
	return s.migrate(ctx, migrations)
}

func (s *Store) migrate(ctx context.Context, files fs.FS) error {
	if _, err := s.db.ExecContext(ctx, schemaVersion); err != nil {
		return fmt.Errorf("create schema version: %w", err)
	}

	root := "migrations"
	entries, err := fs.ReadDir(files, root)
	if err != nil {
		return fmt.Errorf("open migrations directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := path.Join(root, entry.Name())
		sql, err := fs.ReadFile(files, path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		if err := s.applyMigration(ctx, path, string(sql)); err != nil {
			return fmt.Errorf("apply migration %s: %w", path, err)
		}
	}

	return nil
}

func (s *Store) applyMigration(ctx context.Context, path, query string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	hash := sha256.Sum256([]byte(query))
	res, err := tx.ExecContext(ctx,
		"INSERT INTO schema_version(migration, sha256) VALUES(?, ?) ON CONFLICT(migration) DO NOTHING",
		path, hash[:],
	)
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check migration insert: %w", err)
	}
	// ie 0 rows inserted, therefore there's a entry in schema_version for this migration
	if rows == 0 {
		var storedHash []byte
		if err := tx.QueryRowContext(ctx,
			"SELECT sha256 FROM schema_version WHERE migration = ?", path,
		).Scan(&storedHash); err != nil {
			return fmt.Errorf("check schema status: %w", err)
		}
		if !bytes.Equal(storedHash, hash[:]) {
			return fmt.Errorf("hash mismatch")
		}
		return nil
	}

	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("apply sql: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database migration: %w", err)
	}

	return nil
}
