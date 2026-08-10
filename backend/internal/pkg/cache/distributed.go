package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedCache provides distributed caching with cache-aside pattern
type DistributedCache struct {
	client     *redis.Client
	localCache sync.Map // In-memory L1 cache
	localTTL   time.Duration
}

// NewDistributedCache creates a new distributed cache
func NewDistributedCache(client *redis.Client, localTTL time.Duration) *DistributedCache {
	return &DistributedCache{
		client:   client,
		localTTL: localTTL,
	}
}

// Get retrieves a value with L1 (local) + L2 (Redis) cache strategy
func (dc *DistributedCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Try L1 cache first
	if val, ok := dc.localCache.Load(key); ok {
		if entry, ok := val.(*cacheEntry); ok {
			if time.Now().Before(entry.expiry) {
				return json.Unmarshal(entry.data, dest)
			}
			dc.localCache.Delete(key)
		}
	}

	// Try L2 cache (Redis)
	data, err := dc.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss")
		}
		return err
	}

	// Store in L1 cache
	dc.localCache.Store(key, &cacheEntry{
		data:   data,
		expiry: time.Now().Add(dc.localTTL),
	})

	return json.Unmarshal(data, dest)
}

// Set stores a value in both L1 and L2 caches
func (dc *DistributedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Store in L2 cache (Redis)
	if err := dc.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return err
	}

	// Store in L1 cache with shorter TTL
	localTTL := dc.localTTL
	if localTTL > ttl {
		localTTL = ttl
	}
	dc.localCache.Store(key, &cacheEntry{
		data:   data,
		expiry: time.Now().Add(localTTL),
	})

	return nil
}

// Delete removes a value from both L1 and L2 caches
func (dc *DistributedCache) Delete(ctx context.Context, key string) error {
	dc.localCache.Delete(key)
	return dc.client.Del(ctx, key).Err()
}

// InvalidatePattern invalidates all keys matching a pattern
func (dc *DistributedCache) InvalidatePattern(ctx context.Context, pattern string) error {
	// Clear local cache
	dc.localCache.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			// Simple pattern matching
			if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
				prefix := pattern[:len(pattern)-1]
				if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
					dc.localCache.Delete(key)
				}
			}
		}
		return true
	})

	// Clear Redis keys
	iter := dc.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		dc.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

// cacheEntry represents a local cache entry
type cacheEntry struct {
	data   []byte
	expiry time.Time
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Sets      int64
	Deletes   int64
	Evictions int64
}

// GetStats returns cache statistics
func (dc *DistributedCache) GetStats() CacheStats {
	// In production, use atomic counters
	return CacheStats{}
}

// Lock acquires a distributed lock using Redis
func (dc *DistributedCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return dc.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
}

// Unlock releases a distributed lock
func (dc *DistributedCache) Unlock(ctx context.Context, key string) error {
	return dc.client.Del(ctx, "lock:"+key).Err()
}

// IsHealthy checks if the cache is healthy
func (dc *DistributedCache) IsHealthy(ctx context.Context) error {
	return dc.client.Ping(ctx).Err()
}
