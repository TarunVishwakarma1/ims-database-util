package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisInstance *redis.Client
	redisOnce     sync.Once
)

func InitRedis(ctx context.Context, connString string) (*redis.Client, error) {
	var err error

	redisOnce.Do(func() {
		opts, parseErr := redis.ParseURL(connString)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse redis URL: %w", parseErr)
			return
		}

		opts.DialTimeout = 500 * time.Millisecond
		opts.ReadTimeout = 500 * time.Millisecond
		opts.WriteTimeout = 500 * time.Millisecond
		opts.PoolSize = 100

		client := redis.NewClient(opts)

		if pingErr := client.Ping(ctx); pingErr != nil {
			err = fmt.Errorf("redis is not reachable: %v", pingErr)
			return
		}

		slog.Info("✅ Successfully connected to Redis")
		redisInstance = client
	})
	return redisInstance, err
}

func GetRedis() *redis.Client {
	if redisInstance == nil {
		panic("Redis client accessed before initialization")
	}
	return redisInstance
}
