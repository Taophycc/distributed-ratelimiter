package ratelimit

import (
	"sync"
)

type Limiter struct {
	mu sync.Mutex
	buckets map[string]Bucket
	newBucket func(key string) Bucket // factory method pattern - creates a new bucket instance , so that the code using it (Limiter) doesn't need to know the concrete details of construction i.e bucket can be created with different parameters or even different types of buckets (TokenBucket, FixedWindow, SlidingWindow) without changing the Limiter code.
}

func NewLimiter(newBucket func(key string) Bucket) *Limiter {
	return &Limiter{
		buckets: make(map[string]Bucket),
		newBucket: newBucket,
	}
}

func (l *Limiter) Allow(key string) Result {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = l.newBucket(key)
		l.buckets[key] = bucket
	}
	l.mu.Unlock()

	return bucket.Allow()

}