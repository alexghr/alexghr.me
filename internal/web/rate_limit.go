package web

import (
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"
)

type leakyBucket struct {
	mu       sync.Mutex
	rps      float64
	avail    float64
	capacity float64
	last     time.Time
	logger   *slog.Logger
}

func newLeakyBucket(logger *slog.Logger, capacity, rate float64) *leakyBucket {
	return &leakyBucket{
		capacity: capacity,
		rps:      rate,
		avail:    capacity,
		last:     time.Now(),
		logger:   logger,
	}
}

func (l *leakyBucket) admit() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.avail = math.Min(l.capacity, l.avail+elapsed*l.rps)
	l.last = now

	if l.avail >= 1 {
		l.avail -= 1
		return true
	}

	return false
}

func (l *leakyBucket) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l.admit() {
			next.ServeHTTP(w, r)
			return
		}

		if l.logger != nil {
			l.logger.WarnContext(r.Context(), "request blocked", "method", r.Method, "path", r.URL.Path)
		}

		w.WriteHeader(http.StatusTooManyRequests)
	})
}
