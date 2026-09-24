package middleware

import (
	"fmt"
	"net/http"
	"time"
	"github.com/Taophycc/ratelimit"
)

type KeyFunc func(*http.Request) string

func RateLimitMiddleware(l *ratelimit.Limiter, keyFn KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)

			w.Header().Set("x-RateLimit-Limit", fmt.Sprintf("%d", l.Limit(key)))
			
			// check if the request is allowed, if not set the appropriate headers and return 429
			if !l.Allow(key) {
				reset := l.ResetAt(key)
				retryAfter := int(time.Until(reset).Seconds())
				if retryAfter < 0 {
					retryAfter = 0
				}
				w.Header().Set("x-RateLimit-Remaining", "0")
				w.Header().Set("x-RateLimit-Reset", fmt.Sprintf("%d", reset.Unix()))
				// retry after header is only set when the limit is exceeded i.e 429
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			// else request allowed to go through, set the remaining and reset headers and call the next handler
			w.Header().Set("x-RateLimit-Remaining", fmt.Sprintf("%d", l.Remaining(key)))
			w.Header().Set("x-RateLimit-Reset", fmt.Sprintf("%d", l.ResetAt(key).Unix()))
			next.ServeHTTP(w, r)
		})
	}
}

