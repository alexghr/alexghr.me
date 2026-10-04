package state

import (
	"context"
	"database/sql"
	"fmt"
)

type Tag struct {
	ID   int64
	Name string
}

func (s *Store) UpsertTag(ctx context.Context, name string) (Tag, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Tag{}, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback()

	tag, err := s.upsertTag(ctx, tx, name)
	if err != nil {
		return Tag{}, err
	}

	if err := tx.Commit(); err != nil {
		return Tag{}, fmt.Errorf("commit tx: %w", err)
	}

	return tag, nil
}

func (s *Store) upsertTag(ctx context.Context, tx *sql.Tx, name string) (Tag, error) {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO tags(name) VALUES(?) ON CONFLICT(name) DO NOTHING`,
		name,
	)

	if err != nil {
		return Tag{}, fmt.Errorf("insert tag: %w", err)
	}

	var id int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM tags where name = ?`, name,
	).Scan(&id)
	if err != nil {
		return Tag{}, fmt.Errorf("get tag id: %w", err)
	}

	return Tag{
		ID:   id,
		Name: name,
	}, nil
}
