package main

import (
	"context"
	"ims-database-util/internal/config"
	"ims-database-util/internal/repository"
	"ims-database-util/internal/router"
	"ims-database-util/internal/storage"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// main boots the HTTP API server, initializes dependencies, and handles graceful shutdown.
//
// It loads application configuration, creates a 10-second startup context, and initializes
// PostgreSQL and Redis connections. It constructs the application router and HTTP server
// (with 10s read/write timeouts and a 120s idle timeout), starts the server, and blocks
// until a SIGINT or SIGTERM is received. Once signaled, it attempts a graceful shutdown
// with a 30-second timeout and logs shutdown outcome.
func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := storage.InitPostgres(ctx, cfg.PostgresURL)
	if err != nil {
		slog.Error("Fatal: Database connection failed:", "Error", err)
	}
	defer pgPool.Close()

	rdb, err := storage.InitRedis(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("Fatal: Redis connection failed:", "Error", err)
	}
	defer rdb.Close()

	userRepo := repository.NewUserRepository(pgPool)

	appRouter := router.Setup(cfg, userRepo)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      appRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("🚀 Server listening on", "PORT", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("Error while Listen and serve:", "ERROR", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown:", "ERROR", err)
	}

	slog.Info("Server exited properly")

}
