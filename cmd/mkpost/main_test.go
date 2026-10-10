package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexghr/alexghr.me/internal/state"
)

func TestRunCreateEditAndForce(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "posts.sqlite3")
	write := func(name, data string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(META_FILE, `{"slug":"test","title":"Test"}`)
	write(POST_FILE, `<p>original</p>`)
	if err := os.Mkdir(filepath.Join(dir, ASSETS_DIR), 0o700); err != nil {
		t.Fatal(err)
	}
	write("assets/test.svg", `<svg xmlns="http://www.w3.org/2000/svg"/>`)
	if err := run(dbPath, dir, false); err != nil {
		t.Fatal(err)
	}
	afterCreate, err := state.Open(context.Background(), dbPath, 1)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := afterCreate.GetPost(context.Background(), "test")
	afterCreate.Close()
	if err != nil || draft.Status != state.PostStatusDraft {
		t.Fatalf("post without status should start as draft: %+v, %v", draft, err)
	}
	write(META_FILE, `{"slug":"test","title":"Test","status":"published"}`)
	write(POST_FILE, `<p>edited</p><img src="assets/test.svg">`)
	if err := run(dbPath, dir, false); err != nil {
		t.Fatal(err)
	}
	write(META_FILE, `{"slug":"test","title":"Renamed","status":"published"}`)
	if err := run(dbPath, dir, false); !errors.Is(err, errPostTitleMismatch) {
		t.Fatalf("expected title guard, got %v", err)
	}
	beforeForce, err := state.Open(context.Background(), dbPath, 1)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := beforeForce.GetPost(context.Background(), "test")
	beforeForce.Close()
	if err != nil || unchanged.ID != draft.ID || unchanged.Status != state.PostStatusPublished || unchanged.Title != "Test" || unchanged.Body != `<p>edited</p><img src="assets/test.svg">` {
		t.Fatalf("rejected rename changed the stored post: %+v, %v", unchanged, err)
	}
	if err := run(dbPath, dir, true); err != nil {
		t.Fatal(err)
	}
	s, err := state.Open(context.Background(), dbPath, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	post, err := s.GetPost(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if post.Title != "Renamed" || post.Status != state.PostStatusPublished || post.Body != `<p>edited</p><img src="assets/test.svg">` || len(post.Assets) != 1 || post.Assets[0].ContentType != "image/svg+xml" {
		t.Fatalf("unexpected imported post: %+v", post)
	}
}
