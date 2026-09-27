package ratelimit

import (
	"time"
)

type Rule struct {
	Capacity int64
	Refill time.Duration
}

type Config struct {
	Default Rule
	PerKey map[string]Rule
}

func (c *Config) RuleFor(key string) Rule {
	if r, ok := c.PerKey[key]; ok {
		return r
	}
	return c.Default
}