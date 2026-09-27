package main

import (
	"log"
	"net/http"
	"time"
	"github.com/Taophycc/ratelimit"
	"github.com/Taophycc/ratelimit/middleware"
)

func main() {
	apiKeyConfig := &ratelimit.Config{
    Default: ratelimit.Rule{Capacity: 1000, Refill: time.Minute},
    PerKey: map[string]ratelimit.Rule{
        "premium-key-1": {Capacity: 10000, Refill: time.Minute},
		"free-tier-key": {Capacity: 50, Refill: time.Second},
    },
	}
	apiKeyLimiter := ratelimit.NewLimiter(func(key string) ratelimit.Bucket {
		rule := apiKeyConfig.RuleFor(key)
    	return ratelimit.NewSlidingWindow(rule.Capacity, rule.Refill)
	})


	ipLimiter := ratelimit.NewLimiter(func(string) ratelimit.Bucket {
        return ratelimit.NewTokenBucket(100, time.Second)
    })


	pathLimiter := ratelimit.NewLimiter(func(string) ratelimit.Bucket {
		return ratelimit.NewFixedWindow(200, time.Minute)
	})
	

	byIPKeyFunc := func(r *http.Request) string {
		return r.RemoteAddr
	}

	byAPIKeyFunc := func(r *http.Request) string {
		return r.Header.Get("X-API-Key")
	}

	byPathKeyFunc := func(r *http.Request) string { return r.URL.Path }

	requireAPIKey := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") == "" {
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", helloHandler)

	handler := requireAPIKey(
		middleware.RateLimitMiddleware(ipLimiter, byIPKeyFunc)(
			middleware.RateLimitMiddleware(apiKeyLimiter, byAPIKeyFunc)(
				middleware.RateLimitMiddleware(pathLimiter, byPathKeyFunc)(mux),
			),
		),
	)	
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

var helloHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
})