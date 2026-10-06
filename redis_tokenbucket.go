package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed tokenbucket.lua
var tokenBucketScriptSource string

var tokenBucketScript = redis.NewScript(tokenBucketScriptSource)

type RedisTokenBucket struct {
	client   *redis.Client
	key      string
	capacity int64
	interval time.Duration
	timeout  time.Duration
	fallback Bucket
}

func RedisHealthCheck(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}

func NewRedisTokenBucket(client *redis.Client, key string, capacity int64, interval time.Duration, fallback Bucket) *RedisTokenBucket {
	if capacity <= 0 {
        capacity = 1
    }
    if interval < time.Millisecond {
        interval = time.Millisecond
    }
    if fallback == nil {
        fallback = NewTokenBucket(capacity, interval)
    }

	return &RedisTokenBucket{
		client:   client,
		key:      key,
		capacity: capacity,
		interval: interval,
		timeout:  time.Second,
		fallback: fallback,
	}
}

func (r *RedisTokenBucket) Allow() Result {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	res, err := tokenBucketScript.Run(ctx, r.client, []string{r.key},
		r.capacity, max(int64(1),r.interval.Milliseconds())).Result()

	if err != nil {
		return r.fallback.Allow() // Redis down or too slow — fallback to in-memory store
	}

	vals := res.([]interface{})
	allowed := vals[0].(int64) == 1
	remaining := vals[1].(int64)
	resetMs := vals[2].(int64)

	return Result{
		Allowed:   allowed,
		Limit:     r.capacity,
		Remaining: remaining,
		ResetAt:   time.UnixMilli(resetMs),
	}
}