package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alexghr/alexghr.me/internal/state"
)

const maxAssetSize = 10 << 20 // 10 MiB

type adminAsset struct {
	Name string
	URL  string
}

type adminPost struct {
	ID     int64
	Title  string
	Slug   string
	URL    string
	Status state.PostStatus
}

type adminPageModel struct {
	model
	Posts  []adminPost
	Assets []adminAsset
	Stats  state.ViewStats
}

func adminHandler(h handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.serveAdmin)
	mux.HandleFunc("POST /api/upload", h.serveUpload)
	mux.Handle("POST /api/posts/{id}/status", http.NewCrossOriginProtection().Handler(http.HandlerFunc(h.serveUpdatePostStatus)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// make sure this doesn't end up read by bots
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		mux.ServeHTTP(w, r)
	})
}

func (h handler) serveAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := h.templates.get("admin")

	if err != nil {
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	assets, err := h.store.ListAssets(r.Context())
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "list assets", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	posts, err := h.store.ListPosts(r.Context(), state.ListPostsOptions{SortColumn: "created_at"})
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "list posts", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	stats, err := h.store.GetViewStats(r.Context(), now.Add(-7*24*time.Hour), now)
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "load view stats", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}
	m := adminPageModel{model: model{InstanceID: h.id, Debug: h.debug}, Stats: stats}

	for _, post := range posts {
		prefix := ""
		if post.Status == state.PostStatusDraft {
			prefix = "/draft"
		}

		publicURL := fmt.Sprintf("%s/blog/%s/", prefix, post.Slug)
		m.Posts = append(m.Posts, adminPost{
			ID:     post.ID,
			Title:  post.Title,
			Slug:   post.Slug,
			Status: post.Status,
			URL:    publicURL,
		})
	}

	for _, asset := range assets {
		publicURL := fmt.Sprintf("/asset/%x/%s", asset.ID, url.PathEscape(asset.Name))
		if asset.ContentType.IsArchive() {
			publicURL = fmt.Sprintf("/asset/%x", asset.ID)
		}

		m.Assets = append(m.Assets, adminAsset{
			Name: asset.Name,
			URL:  publicURL,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", m); err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "serve admin", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}

func (h handler) serveUpdatePostStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.requestLogger(r).ErrorContext(r.Context(), "invalid post id", "error", err)
		h.renderError(w, r, "invalid post id", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := r.ParseForm(); err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "invalid post form", "post_id", id, "error", err)
		h.renderError(w, r, "invalid post form", http.StatusBadRequest)
		return
	}

	status := state.PostStatus(r.PostForm.Get("status"))
	if !status.Valid() {
		h.requestLogger(r).ErrorContext(r.Context(), "invalid post status", "post_id", id, "error", err)
		h.renderError(w, r, "invalid post status", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdatePostStatus(r.Context(), id, status); err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "update post status", "post_id", id, "error", err)
		h.renderError(w, r, "error updating post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h handler) serveUpload(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "expected a multipart file upload", http.StatusBadRequest)
		return
	}

	for {
		part, err := reader.NextPart()
		if err != nil {
			http.Error(w, "missing or invalid file upload", http.StatusBadRequest)
			return
		}
		if part.FormName() != "file" || part.FileName() == "" {
			continue
		}

		data, err := io.ReadAll(io.LimitReader(part, maxAssetSize+1))
		if err != nil {
			http.Error(w, "could not read file", http.StatusBadRequest)
			return
		}
		if len(data) > maxAssetSize {
			http.Error(w, "file exceeds the 10 MiB limit", http.StatusRequestEntityTooLarge)
			return
		}
		contentType := part.Header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}
		_, err = h.store.AddAsset(r.Context(), part.FileName(), state.AssetContentType(contentType), data)
		if err != nil {
			h.requestLogger(r).ErrorContext(r.Context(), "save uploaded asset", "error", err)
			http.Error(w, "could not save file", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
		return
	}
}
