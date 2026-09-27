package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu sync.Mutex
	buckets map[string]Bucket
	capacity int64
	refillInterval time.Duration
	newBucket func() Bucket // factory method pattern - creates a new bucket instance , so that the code using it (Limiter) doesn't need to know the concrete details of construction i.e bucket can be created with different parameters or even different types of buckets (TokenBucket, FixedWindow, SlidingWindow) without changing the Limiter code.
}

func NewLimiter(newBucket func() Bucket) *Limiter {
	return &Limiter{
		buckets: make(map[string]Bucket),
		newBucket: newBucket,
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = l.newBucket()
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Allow()
}

func (l *Limiter) Remaining(key string) int64 {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = l.newBucket()
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Remaining()
}

func (l *Limiter) ResetAt(key string) time.Time {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = l.newBucket()
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.ResetAt()
}

func (l *Limiter) Limit(key string) int64 {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = l.newBucket()
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Limit()
}