package ratelimit

import (
	"sync"
	"time"
)

type FixedWindow struct {
	mu sync.Mutex
	limit int64
	window time.Duration
	count int64
	windowStart time.Time
}

func NewFixedWindow(limit int64, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit: limit,
		window: window,
		count: 0,
		windowStart: time.Now(),
	}
}

func (fw *FixedWindow) Allow() Result {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	if now.Sub(fw.windowStart) >= fw.window {
		fw.windowStart = now
		fw.count = 0
	}
	allowed := false
	if fw.count < fw.limit {
		fw.count++
		allowed = true
	}

    remaining := fw.limit - fw.count

	resetAt := time.Now()
	if fw.count >=fw.limit {
		resetAt = fw.windowStart.Add(fw.window)
	}

	return Result {
		Allowed: allowed,
		Remaining: remaining,
		ResetAt: resetAt,
		Limit: fw.limit,
	}
	
}
