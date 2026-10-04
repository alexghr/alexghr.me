package web

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"path"
	"time"
)

func (h handler) serveAsset(w http.ResponseWriter, r *http.Request) {
	id, err := hex.DecodeString(r.PathValue("hash"))
	if err != nil || len(id) != sha256.Size {
		http.NotFound(w, r)
		return
	}
	// TODO: loads everythin in memory, consider streaming if asset size grows
	asset, err := h.store.GetAsset(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read asset", "error", err)
		http.Error(w, "could not read asset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", asset.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, asset.Name, time.Time{}, bytes.NewReader(asset.Data))
}

func (h handler) servePostAsset(w http.ResponseWriter, r *http.Request) {
	localPath := "assets/" + r.PathValue("path")
	asset, err := h.store.GetPostAsset(r.Context(), r.PathValue("slug"), localPath)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read post asset", "error", err)
		http.Error(w, "could not read asset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", asset.ContentType)
	// cache these for 5m. Not as long as the immutable URLs but still fine
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("ETag", `"`+hex.EncodeToString(asset.ID)+`"`)
	http.ServeContent(w, r, path.Base(localPath), time.Time{}, bytes.NewReader(asset.Data))
}
