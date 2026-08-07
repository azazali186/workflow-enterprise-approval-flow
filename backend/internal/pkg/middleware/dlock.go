package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DistributedLock represents a distributed lock
type DistributedLock struct {
	redis    *cache.Redis
	key      string
	value    string
	ttl      time.Duration
	acquired bool
}

// NewDistributedLock creates a new distributed lock
func NewDistributedLock(redis *cache.Redis, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		redis: redis,
		key:   "dlock:" + key,
		value: uuid.New().String(),
		ttl:   ttl,
	}
}

// Acquire tries to acquire the lock
func (dl *DistributedLock) Acquire(ctx context.Context) (bool, error) {
	result, err := dl.redis.Client.SetNX(ctx, dl.key, dl.value, dl.ttl).Result()
	if err != nil {
		return false, err
	}
	dl.acquired = result
	return result, nil
}

// Release releases the lock (only if we own it)
func (dl *DistributedLock) Release(ctx context.Context) error {
	if !dl.acquired {
		return nil
	}

	// Use Lua script for atomic check-and-delete
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := dl.redis.Client.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	dl.acquired = false
	return err
}

// IsAcquired returns whether the lock is acquired
func (dl *DistributedLock) IsAcquired() bool {
	return dl.acquired
}

// DistributedLockMiddleware creates a middleware that uses distributed locking
func DistributedLockMiddleware(redis *cache.Redis, cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Only lock on mutating operations
		method := string(c.Request.Method())
		if method != "POST" && method != "PATCH" && method != "DELETE" {
			c.Next(ctx)
			return
		}

		// Create lock key from request path
		path := string(c.Request.URI().Path())
		lockKey := fmt.Sprintf("request:%s:%s", method, path)

		lock := NewDistributedLock(redis, lockKey, 10*time.Second)
		acquired, err := lock.Acquire(ctx)
		if err != nil {
			cfg.Error("failed to acquire lock", zap.Error(err))
			c.JSON(consts.StatusInternalServerError, map[string]string{
				"error": "failed to acquire lock",
			})
			c.Abort()
			return
		}

		if !acquired {
			c.JSON(consts.StatusConflict, map[string]string{
				"error": "request is being processed by another instance",
			})
			c.Abort()
			return
		}

		defer lock.Release(ctx)

		c.Next(ctx)
	}
}

// WithLock executes a function with a distributed lock
func WithLock(ctx context.Context, redis *cache.Redis, key string, ttl time.Duration, fn func() error) error {
	lock := NewDistributedLock(redis, key, ttl)
	acquired, err := lock.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return fmt.Errorf("lock already held by another instance")
	}

	defer lock.Release(ctx)

	return fn()
}
