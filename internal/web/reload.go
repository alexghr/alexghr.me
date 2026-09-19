package web

import (
	"io"
	"net/http"
	"time"
)

func (h handler) notifyReload(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	_, err := io.WriteString(w, ": connected\n\n")
	if err != nil {
		h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			msg := "event: reload\ndata:\n\n"
			_, err := io.WriteString(w, msg)
			if err != nil {
				h.log.ErrorContext(r.Context(), "SSE", "method", r.Method, "path", r.URL.Path, "error", err)
				return
			}

			flusher.Flush()
		case _ = <-r.Context().Done():
			return
		}
	}
}
