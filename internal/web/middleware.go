package web

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func cache(next http.Handler, duration time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Debug != "true" {
			sec := int32(duration.Seconds())
			w.Header().Set("Cache-Control", "public, max-age="+strconv.FormatInt(int64(sec), 10))
		}
		next.ServeHTTP(w, r)
	})
}

func compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Accept-Encoding is more complex than a list of encodings. Make this better.
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") ||
			strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		next.ServeHTTP(compressedResponseWriter{
			ResponseWriter: w,
			writer:         gz,
		}, r)
	})
}

// helper to combine a gzip writer with a HTTP response writer
type compressedResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w compressedResponseWriter) WriteHeader(status int) {
	// clear the content length before sending the header.
	// this is to cover the case where the middleware may have manually set a content length
	// but that value will be wrong after passing through compression
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}

func (w compressedResponseWriter) Write(p []byte) (int, error) {
	// same as WriteHeader
	w.Header().Del("Content-Length")
	return w.writer.Write(p)
}
