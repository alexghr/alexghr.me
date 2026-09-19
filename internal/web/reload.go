package web

import (
	"io"
	"net/http"
)

func (h handler) notifyReload(w http.ResponseWriter, r *http.Request) {
	if !isDebug() {
		w.WriteHeader(http.StatusOK)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("instance_id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
		return
	}
	flusher.Flush()

	reload := func() {
		msg := "event: reload\ndata: 1\n\n"
		if _, err := io.WriteString(w, msg); err != nil {
			h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
			return
		}

		flusher.Flush()
	}

	if id != "" && id != h.id {
		h.log.Debug("Reloading because of instance ID mismatch", "client instance id", id, "server instance id", h.id)
		reload()
	}

	<-r.Context().Done()
}
