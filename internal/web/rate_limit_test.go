package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"
)

func TestLeakyBucketBurst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bucket := newLeakyBucket(nil, 3, 1)
		for i := range 3 {
			if !bucket.admit() {
				t.Fatalf("request %d rejected within burst capacity", i+1)
			}
		}
		if bucket.admit() {
			t.Fatal("request admitted after burst was exhausted")
		}
	})
}

func TestLeakyBucketRefill(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bucket := newLeakyBucket(nil, 1, 1)
		if !bucket.admit() {
			t.Fatal("initial request rejected")
		}

		time.Sleep(500 * time.Millisecond)
		if bucket.admit() {
			t.Fatal("request admitted before a full token refilled")
		}
		if bucket.admit() {
			t.Fatal("rejected request incorrectly added refill time")
		}

		time.Sleep(500 * time.Millisecond)
		if !bucket.admit() {
			t.Fatal("request rejected after a full token refilled")
		}
		if bucket.admit() {
			t.Fatal("refilled token was not consumed")
		}
	})
}

func TestLeakyBucketWrap(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bucket := newLeakyBucket(nil, 1, 1)
		calls := 0
		handler := bucket.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(http.StatusNoContent)
		}))

		for _, want := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != want {
				t.Fatalf("status = %d, want %d", response.Code, want)
			}
		}
		if calls != 1 {
			t.Fatalf("next handler called %d times, want 1", calls)
		}
	})
}
