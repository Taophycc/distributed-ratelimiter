package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucket_Burst(t *testing.T) {
	tb := NewTokenBucket(5, time.Second)

	// Allow 5 requests
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// The 6th request should be denied
	if tb.Allow() {
		t.Error("Expected 6th request to be denied")
	}

}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(5, 50*time.Millisecond)
	tb.Allow() // Consume 1 token

	// Wait for 60 milliseconds to allow for refill
	time.Sleep(60 * time.Millisecond)

	// Now we should be able to allow 5 more requests
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Errorf("Expected request %d to be allowed after refill", i+1)
		}
	}

	// The 6th request should be denied again
	if tb.Allow() {
		t.Error("Expected 6th request to be denied after refill")
	}
}