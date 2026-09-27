package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu sync.Mutex
	capacity int64
	tokens int64
	refillInterval time.Duration
	lastRefill time.Time
}

func NewTokenBucket(capacity int64, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens: capacity,
		refillInterval: refillInterval,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	if elapsed < tb.refillInterval {
		return
	}
	earned := int64(elapsed / tb.refillInterval)
	tb.tokens += earned
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = tb.lastRefill.Add(time.Duration(earned) * tb.refillInterval)
}

func (tb *TokenBucket) Remaining() int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

func (tb *TokenBucket) ResetAt() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()

	// ResetAt represents the next refill boundary, which is always in the future
	// for a valid bucket configuration.
	return tb.lastRefill.Add(tb.refillInterval)
}

func (tb *TokenBucket) Limit() int64 {
	return tb.capacity
}