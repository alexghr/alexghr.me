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

// GetPostAsset reads an asset by its path relative to the post, such as assets/image.png.
func (s *Store) GetPostAsset(ctx context.Context, slug, localPath string) (Asset, PostStatus, error) {
	var asset Asset
	var status PostStatus
	err := s.db.QueryRowContext(ctx, `SELECT a.id, a.name, a.content_type, a.data, p.status
		FROM posts p
		JOIN post_assets pa ON pa.post_id = p.id
		JOIN assets a ON a.id = pa.asset_id
		WHERE p.slug = ? AND pa.local_path = ?`, slug, localPath,
	).Scan(&asset.ID, &asset.Name, &asset.ContentType, &asset.Data, &status)
	if err != nil {
		return Asset{}, "", fmt.Errorf("read post asset: %w", err)
	}
	return asset, status, nil
}
