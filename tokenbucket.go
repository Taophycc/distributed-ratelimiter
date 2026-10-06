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

func (tb *TokenBucket) Allow() Result {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	allowed := false
	if tb.tokens > 0 {
		tb.tokens--
		allowed = true
	}

	resetAt := tb.lastRefill.Add(tb.refillInterval)
    if resetAt.Before(time.Now()) {
        resetAt = time.Now().Add(tb.refillInterval)
    }
	
	return Result{
		Allowed:   allowed,
		Limit:     tb.capacity,
		Remaining: tb.tokens,
		ResetAt:   resetAt,
	}
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
