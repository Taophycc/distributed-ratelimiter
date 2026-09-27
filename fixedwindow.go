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

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	if now.Sub(fw.windowStart) >= fw.window {
		fw.windowStart = now
		fw.count = 0
	}
	if fw.count < fw.limit {
		fw.count++
		return true
	}
	return false
}

func (fw *FixedWindow) Remaining() int64 {
    fw.mu.Lock()
    defer fw.mu.Unlock()
    now := time.Now()
    if now.Sub(fw.windowStart) >= fw.window {
        return fw.limit // window would reset on next call, full budget available
    }
    return fw.limit - fw.count
}

func (fw *FixedWindow) ResetAt() time.Time {
    fw.mu.Lock()
    defer fw.mu.Unlock()
    return fw.windowStart.Add(fw.window)
}

func (fw *FixedWindow) Limit() int64 {
    return fw.limit
}