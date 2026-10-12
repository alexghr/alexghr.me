package web

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func basicAuth(next http.Handler, realm, expectedUser string, hashedPassword []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pw, ok := r.BasicAuth()
		authHeader := fmt.Sprintf("Basic realm=\"%s\"", realm)

		if !ok {
			w.Header().Add("WWW-Authenticate", authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(pw)); err != nil || (expectedUser != "" && user != expectedUser) {
			w.Header().Add("WWW-Authenticate", authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HashPassword(pw []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return base64.URLEncoding.EncodeToString(hash), nil
}

func decodePassword(pw string) ([]byte, error) {
	bytes, err := base64.URLEncoding.DecodeString(pw)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func cache(next http.Handler, duration time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sec := int32(duration.Seconds())
		w.Header().Set("Cache-Control", "public, max-age="+strconv.FormatInt(int64(sec), 10))
		next.ServeHTTP(w, r)
	})
}

func compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")

		encodings, err := parseQValues(r.Header.Get("Accept-Encoding"))
		if err != nil ||
			!encodings.accepts("gzip") ||
			strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		if r.Method == http.MethodHead {
			head := &headResponseWriter{w, false}
			defer head.Close()
			next.ServeHTTP(head, r)
			return
		}

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

type headResponseWriter struct {
	http.ResponseWriter
	headerSent bool
}

func (w *headResponseWriter) WriteHeader(status int) {
	if w.headerSent {
		return
	}

	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
	w.headerSent = true
}

func (w *headResponseWriter) Write(p []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	// send to /dev/null
	return len(p), nil
}

func (w *headResponseWriter) Close() {
	w.WriteHeader(http.StatusOK)
}

type qvalue struct {
	weight float32
	value  string
}

type qvalues struct {
	values []qvalue
}

func parseQValues(header string) (qvalues, error) {
	values := make([]qvalue, 0)

	for item := range strings.SplitSeq(strings.ToLower(header), ",") {
		var weight float32
		value, weightStr, found := strings.Cut(strings.TrimSpace(item), ";")
		if !found {
			weight = 1.0
		} else {
			q := strings.TrimSpace(weightStr)
			if !strings.HasPrefix(q, "q=") {
				return qvalues{}, fmt.Errorf("invalid weight: %s", item)
			}

			parsed, err := strconv.ParseFloat(q[2:], 32)
			if err != nil {
				return qvalues{}, fmt.Errorf("invalid weight %s: %w", item, err)
			}
			if parsed < 0 || parsed > 1 || math.IsNaN(parsed) {
				return qvalues{}, fmt.Errorf("weight out of range %s", item)
			}
			weight = float32(parsed)
		}
		values = append(values, qvalue{value: strings.TrimSpace(value), weight: weight})
	}

	return qvalues{values: values}, nil
}

func (q qvalues) accepts(value string) bool {
	for _, entry := range q.values {
		if entry.value == strings.ToLower(value) {
			return entry.weight > 0
		}
	}

	for _, entry := range q.values {
		if entry.value == "*" && entry.weight > 0 {
			return true
		}
	}

	return false
}
