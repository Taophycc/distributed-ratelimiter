package ratelimit

import (
	"time"
)

type Result struct {
	Allowed bool
	Remaining int64
	Limit int64
	ResetAt time.Time

}

type Bucket interface {
	Allow() Result
}

var _ Bucket = (*TokenBucket)(nil)
var _ Bucket = (*FixedWindow)(nil)
var _ Bucket = (*SlidingWindow)(nil)