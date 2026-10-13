package web

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"time"
)

func (h handler) serveAsset(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	path := r.PathValue("path")

	id, err := hex.DecodeString(hash)
	if err != nil || len(id) != sha256.Size {
		http.NotFound(w, r)
		return
	}

	if archive, ok := h.archives.Get(hash); ok && path != "" && path != "/" {
		h.serveArchiveAsset(w, r, archive)
		return
	}

	// TODO: loads everythin in memory, consider streaming if asset size grows
	asset, err := h.store.GetAsset(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "read asset", "error", err)
		http.Error(w, "could not read asset", http.StatusInternalServerError)
		return
	}

	if asset.ContentType.IsArchive() && path != "" && path != "/" {
		archive, err := zip.NewReader(bytes.NewReader(asset.Data), int64(len(asset.Data)))
		if err != nil {
			h.requestLogger(r).ErrorContext(r.Context(), "read archive", "error", err)
			http.Error(w, "could not read archive", http.StatusInternalServerError)
			return
		}

		h.archives.Set(hash, archive)
		h.serveArchiveAsset(w, r, archive)
		return
	}

	w.Header().Set("Content-Type", asset.ContentType.String())
	if asset.ContentType.IsArchive() {
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": asset.Name,
		}))
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, asset.Name, time.Time{}, bytes.NewReader(asset.Data))
}

// func (h handler) servePostAsset(w http.ResponseWriter, r *http.Request) {
func (h handler) servePostAsset(onlyPublic bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localPath := "assets/" + r.PathValue("path")
		asset, postStatus, err := h.store.GetPostAsset(r.Context(), r.PathValue("slug"), localPath)
		if errors.Is(err, sql.ErrNoRows) {
			h.redirect(w, r)
			return
		}
		if err != nil {
			h.requestLogger(r).ErrorContext(r.Context(), "read post asset", "error", err)
			http.Error(w, "could not read asset", http.StatusInternalServerError)
			return
		}

		if onlyPublic && !postStatus.Public() {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", asset.ContentType.String())

		if onlyPublic && postStatus.Public() {
			// cache these for 5m. Not as long as the immutable URLs but still fine
			w.Header().Set("Cache-Control", "public, max-age=300")
			w.Header().Set("ETag", `"`+hex.EncodeToString(asset.ID)+`"`)
		}

		http.ServeContent(w, r, path.Base(localPath), time.Time{}, bytes.NewReader(asset.Data))
	})
}

func (h handler) serveArchiveAsset(w http.ResponseWriter, r *http.Request, archive *zip.Reader) {
	path := r.PathValue("path")
	if !fs.ValidPath(path) {
		h.renderNotFound(w, r)
		return
	}

	file, err := archive.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		h.renderNotFound(w, r)
		return
	}
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "read archive file", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "stat archive file", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		h.renderNotFound(w, r)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "determine file type", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, path, time.Time{}, bytes.NewReader(data))
}
