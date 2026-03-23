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

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error("No .env file found, relying on evnironment variables ", "Error", err)
	}

	return &Config{
		PostgresURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		Port:        getEnv("PORT", "8081"),
		HMACSecret:  getEnv("HMAC_SECRET", "super-secret-local-dev-key-change-me"),
	}

}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
