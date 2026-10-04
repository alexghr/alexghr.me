package state

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
)

type AssetID []byte

type Asset struct {
	ID          AssetID
	Name        string
	ContentType string
	Data        []byte
}

// Data is going to be empty
func (s *Store) ListAssets(ctx context.Context) ([]Asset, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, content_type FROM assets ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()
	var assets []Asset
	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.ID, &asset.Name, &asset.ContentType); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	return assets, nil
}

func (s *Store) AddAsset(ctx context.Context, name, contentType string, data []byte) (Asset, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Asset{}, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback()

	asset, err := s.addAsset(ctx, tx, name, contentType, data)
	if err != nil {
		return Asset{}, err
	}

	if err := tx.Commit(); err != nil {
		return Asset{}, fmt.Errorf("commit tx: %w", err)
	}

	return asset, nil
}

func (s *Store) addAsset(ctx context.Context, tx *sql.Tx, name, contentType string, data []byte) (Asset, error) {
	hash := sha256.Sum256(data)

	_, err := tx.ExecContext(ctx,
		`INSERT INTO assets(id, name, content_type, data) 
		VALUES(?, ?, ?, ?) 
		ON CONFLICT(id) DO NOTHING`, hash[:], name, contentType, data)

	if err != nil {
		return Asset{}, fmt.Errorf("insert: %w", err)
	}

	return Asset{
		ID:          hash[:],
		Name:        name,
		ContentType: contentType,
		Data:        data,
	}, nil
}

func (s *Store) GetAsset(ctx context.Context, id []byte) (Asset, error) {
	var asset Asset
	err := s.db.QueryRowContext(ctx,
		`SELECT name, content_type, data FROM assets WHERE id=?`, id,
	).Scan(&asset.Name, &asset.ContentType, &asset.Data)

	if err != nil {
		return Asset{}, fmt.Errorf("failed read asset: %w", err)
	}

	return asset, nil
}
