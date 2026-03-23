package main

import (
	"context"
	"ims-database-util/internal/config"
	"ims-database-util/internal/storage"
	"log/slog"
	"time"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := storage.InitPostgres(ctx, cfg.PostgresURL)
	if err != nil {
		slog.Error("Fatal: Database connection failed: %v", "Error", err)
	}
	defer pgPool.Close()

	rdb, err := storage.InitRedis(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("Fatal: Redis connection failed: %v", "Error", err)
	}
	rdb.Close()

	slog.Info("🚀 Server is ready and listening on port", "PORT", cfg.Port)

	select {}
}
