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
	HMACSecret  string
}

// Load attempts to load environment variables from a .env file and returns a *Config
// populated from environment variables with sensible defaults.
//
// If loading a .env file fails, the error is logged and environment variables are used.
// The returned Config fields are sourced from the following environment variables with
// their respective fallbacks:
//   - DATABASE_URL: "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
//   - REDIS_URL:    "redis://localhost:6379/0"
//   - PORT:         "8081"
//   - HMAC_SECRET:  "super-secret-local-dev-key-change-me"
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error("No .env file found, relying on evnironment variables ", "Error", err)
	}

	return &Config{
		PostgresURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		Port:        getEnv("PORT", "8081"),
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
