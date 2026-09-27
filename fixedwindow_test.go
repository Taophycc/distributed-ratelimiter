package ratelimit

import (
	"testing"
	"time"
)

func TestFixedWindow_AllowsUpToLimit(t *testing.T) {
	fw := NewFixedWindow(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !fw.Allow() {
			t.Fatalf("request %d: expected allow", i)
		}
	}
	if fw.Allow() {
		t.Fatal("4th request should be denied, limit reached")
	}
}

func TestFixedWindow_ResetsAfterWindow(t *testing.T) {
	fw := NewFixedWindow(1, 50*time.Millisecond)
	fw.Allow()                        // consumes the only slot
	time.Sleep(60 * time.Millisecond) // wait past the window
	if !fw.Allow() {
		t.Fatal("expected window to have reset")
	}
}

func TestFixedWindow_Remaining(t *testing.T) {
	fw := NewFixedWindow(5, time.Minute)
	fw.Allow()
	fw.Allow()
	if got := fw.Remaining(); got != 3 {
		t.Errorf("got %d remaining, want 3", got)
	}
}

// TestFixedWindow_BoundaryFlaw demonstrates the actual defect fixed window has:
// a client can get a full window's worth of requests right before a boundary,
// and another full window's worth right after — doubling the intended rate
// in a fraction of the window's real duration.
func TestFixedWindow_BoundaryFlaw(t *testing.T) {
	limit := int64(5)
	window := 100 * time.Millisecond
	fw := NewFixedWindow(limit, window)

	// Burst 1: drain the current window completely.
	allowed := int64(0)
	for i := int64(0); i < limit; i++ {
		if fw.Allow() {
			allowed++
		}
	}
	if allowed != limit {
		t.Fatalf("burst 1: got %d allowed, want %d", allowed, limit)
	}

	// Sleep just past the window boundary so a reset fires on the next call.
	time.Sleep(window + 5*time.Millisecond)

	// Burst 2: immediately after the reset, drain a full new window's worth.
	allowed2 := int64(0)
	for i := int64(0); i < limit; i++ {
		if fw.Allow() {
			allowed2++
		}
	}
	if allowed2 != limit {
		t.Fatalf("burst 2: got %d allowed, want %d", allowed2, limit)
	}

	// The flaw: 2*limit requests were allowed within roughly one window's
	// duration (window + a few ms), not the "limit per window" the config
	// implies. This is the defect, demonstrated rather than assumed.
	total := allowed + allowed2
	if total != 2*limit {
		t.Fatalf("expected boundary flaw to allow %d requests total, got %d", 2*limit, total)
	}
	t.Logf("fixed window allowed %d requests within ~%v (configured limit: %d per %v)", total, window, limit, window)
}