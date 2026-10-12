package web

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alexghr/alexghr.me/internal/state"
)

func TestRedirects(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "redirects.sqlite3")
	store, err := state.Open(ctx, dbPath, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	for _, post := range []struct {
		slug   string
		status state.PostStatus
	}{
		{"existing", state.PostStatusPublished},
		{"draft", state.PostStatusDraft},
		{"archived", state.PostStatusArchived},
	} {
		if _, err := store.AddPost(ctx, state.PostInit{
			PostMetadataInit: state.PostMetadataInit{Slug: post.slug, Title: post.slug, Status: &post.status},
			Body:             "<p>Existing post</p>",
			BodyType:         state.PostBodyHtml,
			Assets:           []state.PostAssetInit{{Path: "assets/new.txt", ContentType: "text/plain", Data: []byte("asset")}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	for _, redir := range []struct {
		from, to          string
		active, permanent bool
	}{
		{"/old/nested/file", "/new/file", true, true},
		{"/temporary", "/new", true, false},
		{"/configured", "/new?mode=download#section", true, true},
		{"/blog/old-post/", "/blog/existing/", true, true},
		{"/blog/existing/assets/old.txt", "/blog/existing/assets/new.txt", true, true},
		{"/blog/existing/", "/new", true, true},
		{"/blog/draft/", "/new", true, true},
		{"/blog/draft/assets/new.txt", "/new", true, true},
		{"/inactive", "/new", false, true},
		{"/invalid", "/%zz", true, true},
	} {
		if _, err := db.Exec("INSERT INTO redirects(from_path, to_url, active, permanent) VALUES(?, ?, ?, ?)",
			redir.from, redir.to, redir.active, redir.permanent); err != nil {
			t.Fatal(err)
		}
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := NewHTTPHandler(store, logger, HandlerOptions{
		RateLimitBurst:  100,
		RateLimitRefill: 1,
		PublicURL:       "https://www.alexghr.me/",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, method, url string
		status            int
		location          string
	}{
		{"missing", "GET", "/missing", 404, ""},
		{"missing post", "GET", "/blog/missing/", 404, ""},
		{"nested permanent", "GET", "/old/nested/file?download=1", 308, "/new/file?download=1"},
		{"head", "HEAD", "/old/nested/file", 308, "/new/file"},
		{"temporary", "GET", "/temporary", 307, "/new"},
		{"incoming query encoding", "GET", "/temporary?token=a%2Fb&tag=x&tag=y", 307, "/new?token=a%2Fb&tag=x&tag=y"},
		{"destination query and fragment", "GET", "/configured", 308, "/new?mode=download#section"},
		{"merged query", "GET", "/configured?utm=a&mode=preview&tag=x&tag=y", 308, "/new?mode=download&tag=x&tag=y&utm=a#section"},
		{"renamed post", "GET", "/blog/old-post/", 308, "/blog/existing/"},
		{"renamed asset", "GET", "/blog/existing/assets/old.txt", 308, "/blog/existing/assets/new.txt"},
		{"existing post wins", "GET", "/blog/existing/", 200, ""},
		{"existing asset", "GET", "/blog/existing/assets/new.txt", 200, ""},
		{"draft stays private", "GET", "/blog/draft/", 404, ""},
		{"draft asset stays private", "GET", "/blog/draft/assets/new.txt", 404, ""},
		{"archived post", "GET", "/blog/archived/", 200, ""},
		{"archived asset", "GET", "/blog/archived/assets/new.txt", 200, ""},
		{"inactive", "GET", "/inactive", 404, ""},
		{"invalid destination", "GET", "/invalid", 404, ""},
		{"unsupported method", "POST", "/temporary", 405, ""},
		{"static route", "GET", "/static/favicon.svg", 200, ""},
		{"admin route", "GET", "/admin/", 401, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.url, nil))
			if rr.Code != tc.status || rr.Header().Get("Location") != tc.location {
				t.Fatalf("got HTTP %d Location=%q; want HTTP %d Location=%q",
					rr.Code, rr.Header().Get("Location"), tc.status, tc.location)
			}
			if tc.method == http.MethodHead && rr.Body.Len() != 0 {
				t.Fatal("HEAD response contains a body")
			}
			if tc.status == http.StatusMethodNotAllowed && rr.Header().Get("Allow") != "GET, HEAD" {
				t.Fatalf("unexpected Allow header: %q", rr.Header().Get("Allow"))
			}
		})
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/temporary", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("database error returned HTTP %d, want 500", rr.Code)
	}
}
