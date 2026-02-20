package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepo implements repository.CacheRepository using Redis.
type CacheRepo struct {
	client *redis.Client
}

func NewCacheRepo(client *redis.Client) *CacheRepo {
	return &CacheRepo{client: client}
}

func (r *CacheRepo) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // cache miss — not an error
	}
	return val, err
}

func (r *CacheRepo) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *CacheRepo) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
