package main

import (
	"context"
	"ims-database-util/internal/app"
	"ims-database-util/internal/config"
	"ims-database-util/internal/logger"
	"ims-database-util/internal/server"
	"ims-database-util/internal/storage"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	slog.SetDefault(logger.NewLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Infrastructure
	pgPool, err := storage.InitPostgres(ctx, cfg.PostgresURL)
	if err != nil {
		slog.Error("Fatal: Database connection failed", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	rdb, err := storage.InitRedis(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("Fatal: Redis connection failed", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// Application (repos → services wired internally)
	application := app.New(cfg, pgPool, rdb)

	// Servers (HTTP + gRPC)
	srv := server.New(application)
	srv.Start()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	srv.Stop(shutdownCtx)
}
