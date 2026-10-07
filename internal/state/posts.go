package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type PostBodyType string

const (
	PostBodyHtml         PostBodyType = "html"
	PostBodyHtmlTemplate PostBodyType = "html-template"
)

type PostMetadataInit struct {
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	PublishedAt *time.Time `json:"publishedAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	Tags        []string   `json:"tags"`
}

type PostAssetInit struct {
	Path        string
	ContentType string
	Data        []byte
}

type PostInit struct {
	PostMetadataInit
	Body     string
	BodyType PostBodyType
	Assets   []PostAssetInit
}

type PostAsset struct {
	ID          AssetID
	Path        string
	ContentType string
	Data        []byte
}

type PostMetadata struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	BodyType    PostBodyType
	CreatedAt   time.Time
	PublishedAt time.Time
	UpdatedAt   time.Time
}

type Post struct {
	PostMetadata
	Body   string
	Tags   []int64
	Assets []PostAsset
}

func (s *Store) UpsertPost(ctx context.Context, post PostInit) (Post, error) {
	return s.savePost(ctx, post, true)
}

func (s *Store) AddPost(ctx context.Context, post PostInit) (Post, error) {
	return s.savePost(ctx, post, false)
}

func (s *Store) savePost(ctx context.Context, post PostInit, upsert bool) (Post, error) {
	if strings.TrimSpace(post.Slug) == "" {
		return Post{}, fmt.Errorf("empty slug")
	}
	if strings.TrimSpace(post.Title) == "" {
		return Post{}, errors.New("empty title")
	}
	if post.BodyType != PostBodyHtml && post.BodyType != PostBodyHtmlTemplate {
		return Post{}, fmt.Errorf("unsupported post body type %q", post.BodyType)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Post{}, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback()

	var existing Post
	if upsert {
		existing, err = s.getPost(ctx, tx, post.Slug)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return Post{}, err
		}
		if err == nil {
			// take a snapshot of the Post
			if err := snapshotPost(ctx, tx, existing.ID); err != nil {
				return Post{}, err
			}
		}
	}

	var publishedAt, updatedAt sql.NullString
	if post.PublishedAt != nil {
		publishedAt = sql.NullString{String: post.PublishedAt.UTC().Format(time.RFC3339Nano), Valid: true}
	}
	if post.UpdatedAt != nil {
		// updated_at could be passed in for posts that got imported from the old site
		updatedAt = sql.NullString{String: post.UpdatedAt.UTC().Format(time.RFC3339Nano), Valid: true}
	}

	var id int64
	if existing.ID != 0 {
		// it's an update so fill in updated_at
		if post.UpdatedAt == nil {
			updatedAt = sql.NullString{String: time.Now().UTC().Format(time.RFC3339Nano), Valid: true}
		}

		_, err = tx.ExecContext(ctx, `UPDATE posts SET title = ?, description = ?, body = ?, body_type = ?,
			published_at = COALESCE(?, published_at), updated_at = ? WHERE id = ?`,
			post.Title, post.Description, post.Body, post.BodyType, publishedAt, updatedAt, existing.ID)
		if err != nil {
			return Post{}, fmt.Errorf("update post: %w", err)
		}

		// delete old tags and assets
		for _, query := range []string{"DELETE FROM post_tags WHERE post_id = ?", "DELETE FROM post_assets WHERE post_id = ?"} {
			if _, err := tx.ExecContext(ctx, query, existing.ID); err != nil {
				return Post{}, fmt.Errorf("replace post associations: %w", err)
			}
		}

		id = existing.ID
	} else {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO posts(slug, title, description, body, body_type, published_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)`,
			post.Slug, post.Title, post.Description, post.Body, post.BodyType,
			publishedAt, updatedAt,
		)

		if err != nil {
			return Post{}, fmt.Errorf("insert post: %w", err)
		}

		id, err = res.LastInsertId()
		if err != nil {
			return Post{}, fmt.Errorf("get inserted post id: %w", err)
		}
	}

	_, err = s.upsertPostTags(ctx, tx, id, post.Tags)
	if err != nil {
		return Post{}, err
	}
	if _, err := s.addPostAssets(ctx, tx, id, post.Assets); err != nil {
		return Post{}, err
	}
	stored, err := s.getPost(ctx, tx, post.Slug)
	if err != nil {
		return Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return Post{}, fmt.Errorf("commit tx: %w", err)
	}

	return stored, nil
}

func snapshotPost(ctx context.Context, tx *sql.Tx, postID int64) error {
	res, err := tx.ExecContext(ctx, `INSERT INTO post_revisions
		(post_id, slug, title, description, body, body_type, created_at, published_at, updated_at)
		SELECT id, slug, title, description, body, body_type, created_at, published_at, updated_at
		FROM posts WHERE id = ?`, postID)
	if err != nil {
		return fmt.Errorf("save post revision: %w", err)
	}
	revisionID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get revision id: %w", err)
	}
	for _, query := range []string{
		`INSERT INTO post_revision_tags(revision_id, tag_id) SELECT ?, tag_id FROM post_tags WHERE post_id = ?`,
		`INSERT INTO post_revision_assets(revision_id, asset_id, local_path) SELECT ?, asset_id, local_path FROM post_assets WHERE post_id = ?`,
	} {
		if _, err := tx.ExecContext(ctx, query, revisionID, postID); err != nil {
			return fmt.Errorf("save revision associations: %w", err)
		}
	}
	return nil
}

func (s *Store) upsertPostTags(ctx context.Context, tx *sql.Tx, postID int64, postTags []string) ([]int64, error) {
	tags := make([]int64, 0, len(postTags))

	for _, tagText := range postTags {
		tag, err := s.upsertTag(ctx, tx, tagText)
		if err != nil {
			return nil, fmt.Errorf("insert tag \"%s\": %w", tagText, err)
		}

		tags = append(tags, tag.ID)
	}

	if len(tags) > 0 {
		var sqlText strings.Builder
		sqlText.WriteString("INSERT INTO post_tags(post_id, tag_id) VALUES")

		bindings := make([]any, 0, 2*len(tags)) // really just an []int64. 2 bindings per tag

		for i, tagID := range tags {
			sqlText.WriteString("(?, ?)")
			if i < len(tags)-1 {
				sqlText.WriteString(",")
			}
			bindings = append(bindings, postID, tagID)
		}

		sqlText.WriteString(" ON CONFLICT(post_id, tag_id) DO NOTHING") // shouldn't have duplicates but let's guard against it

		_, err := tx.ExecContext(ctx, sqlText.String(), bindings...)
		if err != nil {
			return nil, fmt.Errorf("insert post tags: %w", err)
		}
	}

	return tags, nil
}

func (s *Store) addPostAssets(ctx context.Context, tx *sql.Tx, postID int64, postAssets []PostAssetInit) ([]AssetID, error) {
	assets := make([]AssetID, 0, len(postAssets))

	for _, postAsset := range postAssets {
		asset, err := s.addAsset(ctx, tx, filepath.Base(postAsset.Path), postAsset.ContentType, postAsset.Data)
		if err != nil {
			return nil, err
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO post_assets(post_id, asset_id, local_path) VALUES(?, ?, ?)`,
			postID, asset.ID, postAsset.Path,
		); err != nil {
			return nil, fmt.Errorf("insert post asset: %w", err)
		}
		assets = append(assets, asset.ID)
	}

	return assets, nil
}

func (s *Store) ListPosts(ctx context.Context) ([]PostMetadata, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, slug, title, description, body_type,
		created_at, published_at, updated_at FROM posts ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()
	var posts []PostMetadata
	for rows.Next() {
		post, err := scanPostMetadata(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	return posts, nil
}

// GetPost returns a post with its tags and assets. Missing slugs wrap sql.ErrNoRows.
func (s *Store) GetPost(ctx context.Context, slug string) (Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Post{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	post, err := s.getPost(ctx, tx, slug)
	if err != nil {
		return Post{}, err
	}
	if err := tx.Commit(); err != nil {
		return Post{}, fmt.Errorf("commit tx: %w", err)
	}
	return post, nil
}

func (s *Store) getPost(ctx context.Context, tx *sql.Tx, slug string) (Post, error) {
	var post Post

	metadata, err := scanPostMetadata(tx.QueryRowContext(ctx,
		`SELECT id, slug, title, description, body_type, created_at, published_at, updated_at, body
		FROM posts WHERE slug = ?`, slug), &post.Body)
	if err != nil {
		return Post{}, err
	}

	post.PostMetadata = metadata

	rows, err := tx.QueryContext(ctx, `SELECT tag_id FROM post_tags WHERE post_id = ? ORDER BY tag_id`, post.ID)
	if err != nil {
		return Post{}, fmt.Errorf("get post tags: %w", err)
	}

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return Post{}, fmt.Errorf("scan post tag: %w", err)
		}

		post.Tags = append(post.Tags, id)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return Post{}, fmt.Errorf("get post tags: %w", err)
	}

	rows, err = tx.QueryContext(ctx, `SELECT a.id, pa.local_path, a.content_type, a.data
		FROM post_assets pa JOIN assets a ON a.id = pa.asset_id
		WHERE pa.post_id = ? ORDER BY pa.local_path`, post.ID)
	if err != nil {
		return Post{}, fmt.Errorf("get post assets: %w", err)
	}

	for rows.Next() {
		var asset PostAsset
		if err := rows.Scan(&asset.ID, &asset.Path, &asset.ContentType, &asset.Data); err != nil {
			return Post{}, fmt.Errorf("scan post asset: %w", err)
		}
		post.Assets = append(post.Assets, asset)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return Post{}, fmt.Errorf("get post assets: %w", err)
	}

	return post, nil
}

// helper interface for sql.Row and sql.Rows
type scanner interface {
	Scan(...any) error
}

// extra receives any columns selected after the metadata columns.
func scanPostMetadata(row scanner, extra ...any) (PostMetadata, error) {
	var post PostMetadata
	var created string
	var published, updated sql.NullString

	fields := []any{&post.ID, &post.Slug, &post.Title, &post.Description,
		&post.BodyType, &created, &published, &updated}

	if err := row.Scan(append(fields, extra...)...); err != nil {
		return PostMetadata{}, fmt.Errorf("scan post metadata: %w", err)
	}

	var err error
	post.CreatedAt, err = time.Parse(time.RFC3339, created)
	if err != nil {
		return PostMetadata{}, fmt.Errorf("parse created_at: %w", err)
	}

	if published.Valid {
		post.PublishedAt, err = time.Parse(time.RFC3339, published.String)
		if err != nil {
			return PostMetadata{}, fmt.Errorf("parse published_at: %w", err)
		}
	}

	if updated.Valid {
		post.UpdatedAt, err = time.Parse(time.RFC3339, updated.String)
		if err != nil {
			return PostMetadata{}, fmt.Errorf("parse updated_at: %w", err)
		}
	}

	return post, nil
}
