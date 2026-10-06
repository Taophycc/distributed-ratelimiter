package ratelimit

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)



func TestRedisTokenBucket_DistributedConcurrency(t *testing.T) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "localhost:6379"
    }

    client := redis.NewClient(&redis.Options{Addr: addr})
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        t.Skipf("Redis unavailable at %s: %v", addr, err)
    }

    key := fmt.Sprintf("ratelimit:test:%d", time.Now().UnixNano())
    t.Cleanup(func() {
        cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
        defer cleanupCancel()
        _ = client.Del(cleanupCtx, key).Err()
    })

    bucketFactory := func() *RedisTokenBucket {
        return NewRedisTokenBucket(
            client,
            key,
            2,
            time.Hour,
            NewTokenBucket(2, time.Hour),
        )
    }

    var (
        wg       sync.WaitGroup
        allowed  atomic.Int64
        rejected atomic.Int64
    )

    for i := 0; i < 100; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()

            result := bucketFactory().Allow()
            if result.Allowed {
                allowed.Add(1)
            } else {
                rejected.Add(1)
            }
        }()
    }

    wg.Wait()

    if got := allowed.Load(); got != 2 {
        t.Fatalf("allowed requests = %d, want 2", got)
    }
}

func TestRedisTokenBucket_FailOpenUsesFallback(t *testing.T) {
    client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
    defer client.Close()

    fallback := NewTokenBucket(1, time.Minute)
    bucket := NewRedisTokenBucket(client, "ratelimit:fail-open", 1, time.Minute, fallback)

    result := bucket.Allow()
    if !result.Allowed {
        t.Fatal("expected Redis failure to return the fallback result")
    }
    if result.Remaining != 0 {
        t.Fatalf("remaining = %d, want 0", result.Remaining)
    }
}

func TestRedisHealthCheck_FailsWhenRedisIsUnavailable(t *testing.T) {
    client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()

    if err := RedisHealthCheck(ctx, client); err == nil {
        t.Fatal("expected Redis health check to fail for an unreachable server")
    }
}

func TestRedisTokenBucket_SharedLimitAcrossInstances(t *testing.T) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "localhost:6379"
    }

    client := redis.NewClient(&redis.Options{Addr: addr})
    t.Cleanup(func() {
        _ = client.Close()
    })

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        t.Skipf("Redis unavailable at %s: %v", addr, err)
    }

    key := fmt.Sprintf("ratelimit:test:%d", time.Now().UnixNano())
    t.Cleanup(func() {
        cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
        defer cleanupCancel()
        _ = client.Del(cleanupCtx, key).Err()
    })

    first := NewRedisTokenBucket(client, key, 2, time.Hour, NewTokenBucket(2, time.Hour))
    second := NewRedisTokenBucket(client, key, 2, time.Hour, NewTokenBucket(2, time.Hour))

    if result := first.Allow(); !result.Allowed {
        t.Fatal("first request should be allowed")
    }
    if result := second.Allow(); !result.Allowed {
        t.Fatal("second request should be allowed")
    }
    if result := first.Allow(); result.Allowed {
        t.Fatal("third request should be denied by the shared Redis limit")
    }
}