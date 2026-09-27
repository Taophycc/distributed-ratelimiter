package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Taophycc/ratelimit"
	"github.com/Taophycc/ratelimit/middleware"
)

func main() {
	ipLimiter := ratelimit.NewLimiter(func() ratelimit.Bucket {
        return ratelimit.NewTokenBucket(100, time.Second)
    })

    keyLimiter := ratelimit.NewLimiter(func() ratelimit.Bucket {
        return ratelimit.NewFixedWindow(1000, time.Minute) 
    })

	pathLimiter := ratelimit.NewLimiter(func() ratelimit.Bucket {
		return ratelimit.NewSlidingWindow(200, time.Minute)
	})
	

	byIPKeyFunc := func(r *http.Request) string {
		return r.RemoteAddr
	}

	byAPIKeyFunc := func(r *http.Request) string {
		return r.Header.Get("X_API_Key")
	}

	byPathKeyFunc := func(r *http.Request) string { return r.URL.Path }

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", helloHandler)

	handler := middleware.RateLimitMiddleware(ipLimiter, byIPKeyFunc)(
		middleware.RateLimitMiddleware(keyLimiter, byAPIKeyFunc)(
			middleware.RateLimitMiddleware(pathLimiter, byPathKeyFunc)(mux),
		),
	)


	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

var helloHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
})