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

// InitRedis initializes the package-level Redis client using the provided connection URL.
// It parses the URL, configures connection timeouts and pool size, verifies connectivity with Ping,
// and caches the resulting client so subsequent calls return the same instance.
// It returns the initialized client and any error encountered during the first initialization attempt;
// if initialization failed the returned client may be nil.
func InitRedis(ctx context.Context, connString string) (*redis.Client, error) {
	var err error

	redisOnce.Do(func() {
		opts, parseErr := redis.ParseURL(connString)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse redis URL: %w", parseErr)
			return
		}

		opts.DialTimeout = 10000 * time.Millisecond
		opts.ReadTimeout = 500 * time.Millisecond
		opts.WriteTimeout = 500 * time.Millisecond
		opts.PoolSize = 100

		client := redis.NewClient(opts)

		if pingErr := client.Ping(ctx).Err(); pingErr != nil {
			err = fmt.Errorf("redis is not reachable: %w", pingErr)
			return
		}

		slog.Info("✅ Successfully connected to Redis")
		redisInstance = client
	})
	return redisInstance, err
}

// GetRedis returns the package-level Redis client instance.
// It panics if InitRedis has not successfully initialized the client.
func GetRedis() *redis.Client {
	if redisInstance == nil {
		panic("Redis client accessed before initialization")
	}
	return redisInstance
}
