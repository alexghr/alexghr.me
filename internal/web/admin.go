package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/alexghr/alexghr.me/internal/state"
)

const maxAssetSize = 10 << 20 // 10 MiB

func adminHandler(h handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.serveAdmin)
	mux.HandleFunc("POST /api/upload", h.serveUpload)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// make sure this doesn't end up read by bots
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		mux.ServeHTTP(w, r)
	})
}

type adminAsset struct {
	Name string
	URL  string
}

func (h handler) serveAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := h.templates.get("admin")

	if err != nil {
		h.renderError(w, r)
		return
	}

	assets, err := h.store.ListAssets(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list assets", "error", err)
		h.renderError(w, r)
		return
	}
	m := struct {
		model
		Assets []adminAsset
	}{model: model{InstanceID: h.id, Debug: h.debug}}
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
		h.logger.ErrorContext(r.Context(), "serve admin", "method", r.Method, "path", r.URL.Path, "error", err)
	}
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
			h.logger.ErrorContext(r.Context(), "save uploaded asset", "error", err)
			http.Error(w, "could not save file", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
		return
	}
}
