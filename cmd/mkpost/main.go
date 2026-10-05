package main

import (
	"context"
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

func main() {
	dbPath := flag.String("db", "db.sqlite3", "path to the local database")
	slug := flag.String("dir", ".", "directory of post")
	flag.Parse()
	if err := run(*dbPath, *slug); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func run(dbPath, dir string) error {
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

	_, err = store.AddPost(ctx, post)
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
		return state.PostInit{}, fmt.Errorf("load post matadata: %w", err)
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
		BodyType:         "html-template",
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
