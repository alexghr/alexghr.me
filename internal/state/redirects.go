package state

import (
	"context"
	"fmt"
)

type Redirect struct {
	FromPath  string
	ToURL     string
	Active    bool
	Permanent bool
}

func (s *Store) GetRedirect(ctx context.Context, path string) (Redirect, error) {
	var redir Redirect

	err := s.db.QueryRowContext(ctx, `SELECT from_path, to_url, active, permanent FROM redirects WHERE from_path = ?`, path).
		Scan(&redir.FromPath, &redir.ToURL, &redir.Active, &redir.Permanent)

	if err != nil {
		return redir, fmt.Errorf("read redirect: %w", err)
	}

	return redir, nil
}
