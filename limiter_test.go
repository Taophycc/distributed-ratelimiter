package ratelimit

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_ConcurrentDifferentKeys(t *testing.T) {
	l := NewLimiter(100, time.Millisecond)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		key := fmt.Sprintf("user-%d", i)
		go func() {
			defer wg.Done()
			l.Allow(key)
		}()
	}
	wg.Wait()
}

func TestLimiter_LimitUsesPerKeyBucket(t *testing.T) {
	l := NewLimiter(3, time.Second)
	if got := l.Limit("user-a"); got != 3 {
		t.Fatalf("Limit(user-a) = %d, want 3", got)
	}
}

func TestLimiter_ResetAtIsInTheFuture(t *testing.T) {
	l := NewLimiter(3, time.Second)
	if got := l.ResetAt("user-a"); !got.After(time.Now()) {
		t.Fatal("expected next reset time to be in the future")
	}
}