package state

import (
	"context"
	"fmt"
)

type ViewInit struct {
	RequestID string
	Path      string
	UserAgent *string
	Referrer  *string
	Country   *string
}

type View struct {
	ViewInit
	ID     int64
	Status int
}

func (s *Store) InsertView(ctx context.Context, view ViewInit) (*View, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO views(request_id, path, user_agent, referrer, country) VALUES(?, ?, ?, ?, ?)`,
		view.RequestID, view.Path, view.UserAgent, view.Referrer, view.Country)

	if err != nil {
		return nil, fmt.Errorf("insert view: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted view id: %w", err)
	}

	return &View{ViewInit: view, ID: id}, nil
}

func (s *Store) SetStatusCode(ctx context.Context, id int64, status int) error {
	res, err := s.db.ExecContext(ctx, `UPDATE views SET status = ? WHERE id = ? AND status IS NULL`, status, id)
	if err != nil {
		return fmt.Errorf("update view status: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("count rows: %w", err)
	}

	if count != 1 {
		return fmt.Errorf("did not update 1 row id=%d status=%d", id, status)
	}

	return nil
}
