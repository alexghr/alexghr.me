package web

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/alexghr/alexghr.me/internal/state"
)

type requestIdKey struct{}
type requestData struct {
	logger *slog.Logger
	reqID  string
}

func (h handler) initContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := rand.Text()
		logger := h.logger.With("reqID", reqID, "method", r.Method, "path", r.URL.Path)

		data := &requestData{logger, reqID}

		r = r.WithContext(context.WithValue(r.Context(), requestIdKey{}, data))
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r)
	})
}

func (h handler) requestLogger(r *http.Request) *slog.Logger {
	if data, ok := r.Context().Value(requestIdKey{}).(*requestData); ok {
		return data.logger
	}
	return h.logger
}

func (h handler) requestID(r *http.Request) string {
	if data, ok := r.Context().Value(requestIdKey{}).(*requestData); ok {
		return data.reqID
	}
	return ""
}

func (h handler) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isInterestingPage(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		response := &statusCaptureWriter{ResponseWriter: w}
		next.ServeHTTP(response, r)

		status := response.Status()
		requestLogger := h.requestLogger(r)
		requestLogger.InfoContext(r.Context(), "request",
			"status", status,
			"duration", time.Since(start),
		)
	})
}

func (h handler) trackView(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isInterestingPage(r) || r.Method != http.MethodGet || isBot(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		logger := h.requestLogger(r)
		reqID := h.requestID(r)

		var referrer *string
		if u, err := url.Parse(r.Referer()); err == nil && u.Hostname() != "" {
			referrer = optionalString(u.Hostname())
		}

		// trusted header
		country := r.Header.Get("CF-IPCountry")
		if country == "XX" || country == "T1" || len(country) != 2 {
			country = ""
		}
		view, err := h.store.InsertView(r.Context(), state.ViewInit{
			RequestID: reqID,
			Path:      r.URL.Path,
			UserAgent: optionalString(r.UserAgent()),
			Referrer:  referrer,
			Country:   optionalString(country),
		})
		if err != nil {
			logger.WarnContext(r.Context(), "failed to insert page view", "error", err)
		}

		writer := &statusCaptureWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)

		if view != nil {
			// Allow a bounded update even if the client disconnected.
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 2*time.Second)
			defer cancel()
			if err := h.store.SetStatusCode(ctx, view.ID, writer.Status()); err != nil {
				logger.WarnContext(ctx, "failed to set page view status", "error", err)
			}
		}
	})
}

// behold.. The Internet
func isBot(path string) bool {
	path = strings.ToLower(path)
	if strings.Contains(path, ".php") {
		return true
	}
	for _, prefix := range []string{"/wp-admin", "/wp-content", "/wp-includes", "/cgi-bin", "/.git", "/.aws", "/.env"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func isInterestingPage(r *http.Request) bool {
	path := r.URL.Path
	if path == "/rss.xml" || path == "/robots.txt" || path == "/favicon.ico" {
		return false
	}

	for _, prefix := range []string{"/static", "/asset", "/assets", "/api", "/admin", "/draft"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return false
		}
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	return slices.Index(parts, "assets") == -1
}

type statusCaptureWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCaptureWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *statusCaptureWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusCaptureWriter) Write(p []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	return w.ResponseWriter.Write(p)
}

func (w *statusCaptureWriter) Flush() {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
