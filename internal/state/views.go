package state

import (
	"context"
	"fmt"
	"time"
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

type ViewCount struct {
	Value string
	Count int
}

type ViewStats struct {
	Pages, Referrers, Countries, Errors []ViewCount
}

func (s *Store) GetViewStats(ctx context.Context, since, until time.Time) (ViewStats, error) {
	var stats ViewStats

	groups := []struct {
		column               string
		statusMin, statusMax int
		target               *[]ViewCount
	}{
		{"path", 200, 300, &stats.Pages},
		{"path", 400, 600, &stats.Errors},
		{"referrer", 200, 300, &stats.Referrers},
		{"country", 200, 300, &stats.Countries},
	}

	for _, group := range groups {
		rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(`+group.column+`, ''), COUNT(*)
			FROM views WHERE created_at >= ? AND created_at < ? AND status >= ? AND status < ?
			GROUP BY COALESCE(`+group.column+`, '') ORDER BY COUNT(*) DESC, 1 ASC LIMIT 5`,
			since.UTC().Format(time.RFC3339), until.UTC().Format(time.RFC3339), group.statusMin, group.statusMax)

		if err != nil {
			return ViewStats{}, fmt.Errorf("query %s stats: %w", group.column, err)
		}

		for rows.Next() {
			var count ViewCount
			if err := rows.Scan(&count.Value, &count.Count); err != nil {
				rows.Close()
				return ViewStats{}, fmt.Errorf("scan %s stats: %w", group.column, err)
			}
			*group.target = append(*group.target, count)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return ViewStats{}, fmt.Errorf("read %s stats: %w", group.column, err)
		}
	}
	return stats, nil
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
