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

func(sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(sw.currStart)
	if elapsed >= sw.window {
		windowPassed := int64(elapsed/sw.window)
		if windowPassed == 1 {
			sw.prevCount = sw.currCount
		} else {
			sw.prevCount = 0
		}
		sw.currCount = 0
		sw.currStart = sw.currStart.Add(time.Duration(windowPassed) * sw.window) 
		elapsed = now.Sub(sw.currStart)
	}

	weight := float64(sw.window-elapsed)/float64(sw.window)
	estimate := float64(sw.prevCount)*weight + float64(sw.currCount)

	if estimate < float64(sw.limit) {
		sw.currCount++
		return true
	}
	return false

}

func (sw *SlidingWindow) Remaining() int64 {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(sw.currStart)
	if elapsed >= sw.window {
		windowsPassed := int64(elapsed / sw.window)
		if windowsPassed == 1 {
			sw.prevCount = sw.currCount
		} else {
			sw.prevCount = 0
		}
		sw.currCount = 0
		sw.currStart = sw.currStart.Add(time.Duration(windowsPassed) * sw.window)
		elapsed = now.Sub(sw.currStart)
	}

	weight := float64(sw.window-elapsed) / float64(sw.window)
	estimate := float64(sw.prevCount)*weight + float64(sw.currCount)

	remaining := float64(sw.limit) - estimate
	if remaining < 0 {
		remaining = 0
	}
	return int64(remaining)
}

func (sw *SlidingWindow) ResetAt() time.Time {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	// The window's influence fully fades out once a full window has
	// elapsed since it started — that's the point sliding window's
	// estimate is guaranteed back under any previous pressure.
	return sw.currStart.Add(sw.window)
}

func (sw *SlidingWindow) Limit() int64 {
	return sw.limit
}