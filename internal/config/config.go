package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresURL string
	RedisURL    string
	Port        string
	GRPCPort    string
	HMACSecret  string
}

// Load attempts to load environment variables from a .env file and returns a *Config
// populated from environment variables with sensible defaults.
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error("No .env file found, relying on environment variables", "Error", err)
	}

	return &Config{
		PostgresURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		Port:        getEnv("PORT", "8081"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		HMACSecret:  getEnv("HMAC_SECRET", "super-secret-local-dev-key-change-me"),
	}
}

// getEnv returns the value of the environment variable named by key.
// If the variable is not present, it returns the provided fallback string.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
