package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{client: client}
}

func (r *RedisCache) Set(key string, value interface{}) error {
	err := r.client.Set(ctx, key, value, 0).Err()
	return err
}

func (r *RedisCache) Get(key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	return val, err
}

func (r *RedisCache) Delete(key string) error {
	err := r.client.Del(ctx, key).Err()
	return err
}
