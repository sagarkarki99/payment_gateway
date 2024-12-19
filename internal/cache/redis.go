package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var cache *RedisCache

type RedisCache struct {
	client *redis.Client
}

func InitRedis() error {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "password",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache = &RedisCache{
		client: client,
	}
	return nil
}
