package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func BenchmarkRedisTokenBucket_Allow(b *testing.B) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis unavailable at %s: %v", addr, err)
	}

	key := "ratelimit:bench:redis-bucket"
	b.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
		defer cleanupCancel()
		_ = client.Del(cleanupCtx, key).Err()
	})

	bucket := NewRedisTokenBucket(client, key, 1_000_000, time.Second, nil)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.Allow()
		}
	})
}

func BenchmarkRedisTokenBucket_Allow_Fallback(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:   10 * time.Millisecond,
	})

	bucket := NewRedisTokenBucket(
		client,
		"ratelimit:bench:fallback",
		1_000_000,
		time.Second,
		NewTokenBucket(1_000_000, time.Second),
	)

	bucket.timeout = 10 * time.Millisecond
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}
