package ratelimit

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	mu sync.Mutex
	limit int64
	window time.Duration
	prevCount int64
	currCount int64
	currStart time.Time
}


func NewSlidingWindow(limit int64, window time.Duration) *SlidingWindow {
    if limit <= 0 {
        limit = 1
    }
    if window <= 0 {
        window = time.Second
    }
    return &SlidingWindow{
        limit:     limit,
        window:    window,
        currStart: time.Now(),
    }
}

func(sw *SlidingWindow) Allow() Result {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(sw.currStart)

	if elapsed >= sw.window {
		windowsPassed := int64(elapsed/sw.window)
		if windowsPassed == 1 {
			sw.prevCount = sw.currCount
		} else {
			sw.prevCount = 0
		}
		sw.currCount = 0
		sw.currStart = sw.currStart.Add(time.Duration(windowsPassed) * sw.window) 
		elapsed = now.Sub(sw.currStart)
	}

	weight := float64(sw.window-elapsed)/float64(sw.window)
	estimate := float64(sw.prevCount)*weight + float64(sw.currCount)

	allowed := false
	if estimate < float64(sw.limit) {
		sw.currCount++
		allowed = true
	}

	estimate = float64(sw.prevCount)*weight + float64(sw.currCount)
	remaining := float64(sw.limit) - estimate
	if remaining < 0 {
		remaining = 0
	}

	return Result{
		Allowed: allowed,
		Remaining: int64(remaining),
		Limit: sw.limit,
		ResetAt: sw.currStart.Add(sw.window),
	}

}