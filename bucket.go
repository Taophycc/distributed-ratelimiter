package ratelimit

import (
	"time"
)

type Bucket interface {
	Allow() bool
	Remaining() int64
	ResetAt() time.Time
	Limit() int64
}

var _ Bucket = (*TokenBucket)(nil)
var _ Bucket = (*FixedWindow)(nil)
var _ Bucket = (*SlidingWindow)(nil)