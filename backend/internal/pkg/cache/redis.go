package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(cfg *config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	cfg.Info("redis connection established")
	return &Redis{Client: client}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *Redis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.Client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *Redis) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	return r.Client.HSet(ctx, key, values).Err()
}

func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

func (r *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// SetJSON marshals a value to JSON and stores it in Redis
func (r *Redis) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.Client.Set(ctx, key, data, ttl).Err()
}

// GetJSON retrieves a JSON value from Redis and unmarshals it
func (r *Redis) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// GetWithTTL retrieves a value and its TTL from Redis
func (r *Redis) GetWithTTL(ctx context.Context, key string) (string, time.Duration, error) {
	result := r.Client.Get(ctx, key)
	if result.Err() != nil {
		return "", 0, result.Err()
	}
	ttl, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}
	return result.Val(), ttl, nil
}

// Incr increments a key's value
func (r *Redis) Incr(ctx context.Context, key string) error {
	return r.Client.Incr(ctx, key).Err()
}

// Expire sets a TTL on a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// Ping checks if Redis connection is alive
func (r *Redis) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// TTL returns the remaining TTL for a key
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}
