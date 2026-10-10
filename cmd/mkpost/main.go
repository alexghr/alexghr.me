package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/alexghr/alexghr.me/internal/state"
)

var errPostTitleMismatch = errors.New("post title mismatch")

func main() {
	dbPath := flag.String("db", "db.sqlite3", "path to the local database")
	slug := flag.String("dir", ".", "directory of post")
	force := flag.Bool("force", false, "allow replacing a post with a different title")
	flag.Parse()
	if err := run(*dbPath, *slug, *force); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dbPath, dir string, force bool) error {
	ctx := context.Background()
	store, err := state.Open(ctx, dbPath, 1)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		return err
	}

	fsys, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("open dir %s: %w", dir, err)
	}
	defer fsys.Close()

	post, err := loadPost(fsys.FS())
	if err != nil {
		return fmt.Errorf("open post: %w", err)
	}

	if !force {
		existing, err := store.GetPost(ctx, post.Slug)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check existing post: %w", err)
		}
		if err == nil && existing.Title != post.Title {
			return fmt.Errorf("%w: slug %q has title %q, incoming title %q; use --force to overwrite", errPostTitleMismatch, post.Slug, existing.Title, post.Title)
		}
	}

	_, err = store.UpsertPost(ctx, post)
	if err != nil {
		return fmt.Errorf("store post: %w", err)
	}

	return nil
}

const (
	POST_FILE  = "post.html"
	META_FILE  = "meta.json"
	ASSETS_DIR = "assets"
)

func loadPost(fsys fs.FS) (state.PostInit, error) {
	meta, err := loadPostMetadata(fsys)
	if err != nil {
		return state.PostInit{}, fmt.Errorf("load post metadata: %w", err)
	}

	assets, err := loadPostAssets(fsys)
	if err != nil {
		return state.PostInit{}, fmt.Errorf("load post assets: %w", err)
	}

	body, err := fs.ReadFile(fsys, POST_FILE)
	if err != nil {
		return state.PostInit{}, fmt.Errorf("read post: %w", err)
	}

	return state.PostInit{
		PostMetadataInit: meta,
		Assets:           assets,
		Body:             string(body),
		BodyType:         state.PostBodyHtmlTemplate,
	}, nil
}

func loadPostMetadata(fsys fs.FS) (state.PostMetadataInit, error) {
	file, err := fs.ReadFile(fsys, META_FILE)
	if err != nil {
		return state.PostMetadataInit{}, fmt.Errorf("read metadata file: %w", err)
	}

	var meta state.PostMetadataInit
	if err := json.Unmarshal(file, &meta); err != nil {
		return state.PostMetadataInit{}, fmt.Errorf("parse metadata JSON: %w", err)
	}

	if meta.Slug == "" {
		return state.PostMetadataInit{}, fmt.Errorf("empty slug")
	}

	if meta.Title == "" {
		return state.PostMetadataInit{}, fmt.Errorf("empty title")
	}

	if meta.Status == "" {
		meta.Status = state.PostStatusDraft
	}

	if !meta.Status.Valid() {
		return state.PostMetadataInit{}, fmt.Errorf("invalid status: %s", meta.Status)
	}

	return meta, nil
}

func loadPostAssets(fsys fs.FS) ([]state.PostAssetInit, error) {
	entries, err := fs.ReadDir(fsys, ASSETS_DIR)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("read assets dir: %w", err)
	}

	var assets []state.PostAssetInit
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			data, err := fs.ReadFile(fsys, path.Join(ASSETS_DIR, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("reading assets \"%s\": %w", entry.Name(), err)
			}

			contentType := mime.TypeByExtension(path.Ext(entry.Name()))
			if contentType == "" {
				contentType = http.DetectContentType(data)
			}

			assets = append(assets, state.PostAssetInit{
				Path:        path.Join(ASSETS_DIR, entry.Name()), // use path for web paths
				ContentType: contentType,
				Data:        data,
			})
		}
	}

	return assets, nil
}
