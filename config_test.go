package ratelimit

import (
	"testing"
	"time"
)

func TestConfig_RuleFor_ReturnsPerKeyRule(t *testing.T) {
	cfg := &Config{
		Default: Rule{Capacity: 100, Refill: time.Second},
		PerKey: map[string]Rule{
			"premium-key-1": {Capacity: 1000, Refill: time.Second},
		},
	}

	got := cfg.RuleFor("premium-key-1")
	want := Rule{Capacity: 1000, Refill: time.Second}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestConfig_RuleFor_FallsBackToDefault(t *testing.T) {
	cfg := &Config{
		Default: Rule{Capacity: 100, Refill: time.Second},
		PerKey: map[string]Rule{
			"premium-key-1": {Capacity: 1000, Refill: time.Second},
		},
	}

	got := cfg.RuleFor("some-random-unconfigured-key")
	want := Rule{Capacity: 100, Refill: time.Second}

	if got != want {
		t.Errorf("got %+v, want %+v (expected default)", got, want)
	}
}

func TestConfig_RuleFor_EmptyPerKeyMap(t *testing.T) {
	// Config with no PerKey overrides at all — every key should get Default.
	cfg := &Config{
		Default: Rule{Capacity: 50, Refill: time.Minute},
	}

	got := cfg.RuleFor("anything")
	want := Rule{Capacity: 50, Refill: time.Minute}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestConfig_RuleFor_MultipleDistinctKeys(t *testing.T) {
	cfg := &Config{
		Default: Rule{Capacity: 100, Refill: time.Second},
		PerKey: map[string]Rule{
			"premium-key-1": {Capacity: 1000, Refill: time.Second},
			"free-tier-key": {Capacity: 50, Refill: time.Second},
		},
	}

	tests := []struct {
		name string
		key  string
		want Rule
	}{
		{"premium key", "premium-key-1", Rule{Capacity: 1000, Refill: time.Second}},
		{"free tier key", "free-tier-key", Rule{Capacity: 50, Refill: time.Second}},
		{"unconfigured key", "random-key", Rule{Capacity: 100, Refill: time.Second}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.RuleFor(tt.key)
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}