package ratelimit

import (
	"testing"
	"time"
)

func TestSlidingWindow_AllowsUpToLimit(t *testing.T) {
	sw := NewSlidingWindow(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !sw.Allow().Allowed {
			t.Fatalf("request %d: expected allow", i)
		}
	}
	if sw.Allow().Allowed {
		t.Fatal("4th request should be denied, limit reached")
	}
}

func TestSlidingWindow_FullyRecoversAfterWindow(t *testing.T) {
	sw := NewSlidingWindow(2, 50*time.Millisecond)
	sw.Allow()
	sw.Allow()
	if sw.Allow().Allowed {
		t.Fatal("expected denial, limit reached")
	}

	// wait a full window PLUS a bit — prevCount's weight should have
	// fully decayed to ~0 by now, not just "some" window having passed
	time.Sleep(120 * time.Millisecond)

	if !sw.Allow().Allowed {
		t.Fatal("expected allow after the previous window's influence fully decayed")
	}
}

func TestSlidingWindow_Remaining(t *testing.T) {
	sw := NewSlidingWindow(5, time.Minute)
	sw.Allow()
	result := sw.Allow()
	if got := result.Remaining; got != 3 {
		t.Errorf("got %d remaining, want 3", got)
	}
}

// TestSlidingWindow_NoBoundaryFlaw runs the EXACT same attack pattern as
// TestFixedWindow_BoundaryFlaw — drain a window, wait just past the boundary,
// try to drain a second full window's worth immediately after. FixedWindow
// allows both bursts fully (2*limit total). SlidingWindow should NOT, because
// the previous window's count is still heavily weighted right at the boundary.
func TestSlidingWindow_NoBoundaryFlaw(t *testing.T) {
	limit := int64(5)
	window := 100 * time.Millisecond
	sw := NewSlidingWindow(limit, window)

	// Burst 1: drain the current window completely.
	allowed := int64(0)
	for i := int64(0); i < limit; i++ {
		if sw.Allow().Allowed {
			allowed++
		}
	}
	if allowed != limit {
		t.Fatalf("burst 1: got %d allowed, want %d", allowed, limit)
	}

	// Sleep just past the window boundary — same timing as the fixed
	// window attack. weight should still be close to 1 here, since we're
	// only a hair into the new window.
	time.Sleep(window + 5*time.Millisecond)

	// Burst 2: attempt to drain a full new window's worth immediately.
	allowed2 := int64(0)
	for i := int64(0); i < limit; i++ {
		if sw.Allow().Allowed {
			allowed2++
		}
	}

	total := allowed + allowed2
	t.Logf("sliding window allowed %d total requests within ~%v (configured limit: %d per %v)",
		total, window, limit, window)

	// The whole point: unlike fixed window, this should NOT let both
	// bursts through in full. Some, but not all, of burst 2 should be denied.
	if allowed2 == limit {
		t.Errorf("boundary flaw NOT fixed: burst 2 fully succeeded (%d/%d), same as fixed window's flaw", allowed2, limit)
	}
	if total >= 2*limit {
		t.Errorf("boundary flaw NOT fixed: total allowed (%d) reached fixed window's flawed total (%d)", total, 2*limit)
	}
}