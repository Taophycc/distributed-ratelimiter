package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Taophycc/ratelimit"
	"github.com/Taophycc/ratelimit/middleware"
)

func main() {
	ipLimiter := ratelimit.NewLimiter(100, time.Second)
	keyLimiter := ratelimit.NewLimiter(100, time.Second)

	byIPKeyFunc := func(r *http.Request) string {
		return r.RemoteAddr
	}

	byAPIKeyFunc := func(r *http.Request) string {
		return r.Header.Get("X_API_Key")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", helloHandler)

	handler := middleware.RateLimitMiddleware(ipLimiter, byIPKeyFunc)(
		middleware.RateLimitMiddleware(keyLimiter, byAPIKeyFunc)(mux),
	)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

var helloHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
})