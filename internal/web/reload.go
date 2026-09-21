package web

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (h handler) notifyReload(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("instance_id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	watchers := make([]*fileWatcher, 0, 2)
	dirs := []string{"internal/web/templates", "internal/web/static"}

	for _, dir := range dirs {
		watcher, err := newFileWatcher(dir)
		if err != nil {
			h.log.ErrorContext(r.Context(), "SSE create watcher", "dir", dir, "error", err)
			http.Error(w, "failed to setup watchers", http.StatusInternalServerError)
			return
		}

		watchers = append(watchers, watcher)
	}

	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
		return
	}
	flusher.Flush()

	reload := func() error {
		msg := "event: reload\ndata: 1\n\n"
		if _, err := io.WriteString(w, msg); err != nil {
			return err
		}

		flusher.Flush()
		return nil
	}

	if id != "" && id != h.id {
		h.log.Debug("Reloading because of instance ID mismatch", "client instance id", id, "server instance id", h.id)
		if err := reload(); err != nil {
			h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
			return
		}
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			changed := false
			for _, w := range watchers {
				res, err := w.poll()
				if err != nil {
					h.log.WarnContext(r.Context(), "SSE poll files", "error", err)
					continue
				}

				changed = changed || res
			}

			if changed {
				if err := reload(); err != nil {
					h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
					return
				}
			}
		case <-r.Context().Done():
			return
		}
	}
}

type fileState struct {
	mtime time.Time
}

type fileWatcher struct {
	dir   string
	files map[string]fileState
}

func newFileWatcher(dir string) (*fileWatcher, error) {
	w := &fileWatcher{dir: dir, files: make(map[string]fileState)}
	_, err := w.poll()
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *fileWatcher) poll() (bool, error) {
	dirs := []string{w.dir}
	curr := make(map[string]fileState)

	for len(dirs) > 0 {
		dir := dirs[0]
		dirs = dirs[1:]

		entries, err := os.ReadDir(dir)
		if err != nil {
			return false, fmt.Errorf("open dir %s: %w", dir, err)
		}

		for _, entry := range entries {
			filename := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				dirs = append(dirs, filename)
				continue
			}

			stat, err := os.Stat(filename)
			if err != nil {
				return false, fmt.Errorf("stat file %s: %w", filename, err)
			}

			curr[filename] = fileState{mtime: stat.ModTime()}
		}
	}

	changed := !maps.Equal(curr, w.files)
	w.files = curr

	return changed, nil
}
