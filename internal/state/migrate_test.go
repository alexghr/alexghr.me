package state

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMigrationsApplyCleanly(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx := context.Background()
	store := &Store{db: db}
	files := fstest.MapFS{
		"migrations/0001_create.sql": {Data: []byte("CREATE TABLE records (id INTEGER)")},
	}
	if err := store.migrate(ctx, files); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'records'",
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("records table count = %d, want 1", count)
	}
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM schema_version").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("migration record count = %d, want 1", count)
	}
}

func TestMigrationErrorRollsBack(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx := context.Background()
	store := &Store{db: db}
	files := fstest.MapFS{
		"migrations/0001_bad.sql": {Data: []byte("CREATE TABLE partial (id INTEGER); INVALID SQL")},
	}
	err = store.migrate(ctx, files)
	if err == nil || !strings.Contains(err.Error(), "apply sql") {
		t.Fatalf("migration error = %v, want SQL error", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'partial'",
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("partial table count = %d, want 0", count)
	}
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM schema_version").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("migration record count = %d, want 0", count)
	}
}

func TestMigrationRepeatedWithoutChanges(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx := context.Background()
	store := &Store{db: db}
	files := fstest.MapFS{
		"migrations/0001_create.sql": {Data: []byte("CREATE TABLE records (id INTEGER)")},
	}
	if err := store.migrate(ctx, files); err != nil {
		t.Fatal(err)
	}
	if err := store.migrate(ctx, files); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM schema_version").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("migration record count = %d, want 1", count)
	}
}

func TestMigrationChangedAfterApplying(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx := context.Background()
	store := &Store{db: db}
	files := fstest.MapFS{
		"migrations/0001_create.sql": {Data: []byte("CREATE TABLE original (id INTEGER)")},
	}
	if err := store.migrate(ctx, files); err != nil {
		t.Fatal(err)
	}
	files["migrations/0001_create.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE changed (id INTEGER)")}
	err = store.migrate(ctx, files)
	if err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("changed migration error = %v, want hash mismatch", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'changed'",
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("changed table count = %d, want 0", count)
	}
}
