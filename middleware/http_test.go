package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimit "github.com/Taophycc/distributed-ratelimiter"
)

func TestRateLimit_AllowsUnderCapacity(t *testing.T) {
	limiter := ratelimit.NewLimiter(func(string) ratelimit.Bucket {
		return ratelimit.NewTokenBucket(2, time.Second)
	})
	keyFn := func(r *http.Request) string { return "fixed-key" }

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RateLimitMiddleware(limiter, keyFn)(next)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}
}

func TestRateLimit_DeniesOverCapacity(t *testing.T) {
	limiter := ratelimit.NewLimiter(func(string) ratelimit.Bucket {
		return ratelimit.NewTokenBucket(1, time.Minute)
	})
	keyFn := func(r *http.Request) string { return "fixed-key" }

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RateLimitMiddleware(limiter, keyFn)(next)

	req := httptest.NewRequest("GET", "/", nil)

	// first request consumes the only token
	handler.ServeHTTP(httptest.NewRecorder(), req)

	// second should be denied
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("got status %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestRateLimit_NextNotCalledWhenDenied(t *testing.T) {
	limiter := ratelimit.NewLimiter(func(string) ratelimit.Bucket {
		return ratelimit.NewTokenBucket(0, time.Minute)
	})
	keyFn := func(r *http.Request) string { return "fixed-key" }

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})
	handler := RateLimitMiddleware(limiter, keyFn)(next)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if nextCalled {
		t.Error("next should not be called when request is denied")
	}
}