package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu sync.Mutex
	buckets map[string]*TokenBucket
	capacity int64
	refillInterval time.Duration
}

func NewLimiter(capacity int64, refillInterval time.Duration) *Limiter {
	return &Limiter{
		buckets: make(map[string]*TokenBucket),
		capacity: capacity,
		refillInterval: refillInterval,
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = NewTokenBucket(l.capacity, l.refillInterval)
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Allow()
}

func (l *Limiter) Remaining(key string) int64 {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = NewTokenBucket(l.capacity, l.refillInterval)
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Remaining()
}

func (l *Limiter) ResetAt(key string) time.Time {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = NewTokenBucket(l.capacity, l.refillInterval)
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.ResetAt()
}

func (l *Limiter) Limit(key string) int64 {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = NewTokenBucket(l.capacity, l.refillInterval)
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Limit()
}