package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Port            string
	Env             string // "dev" or "prod"
	LogLevel        string
	ForgeSecretKey  string
	AnthropicAPIKey string
	DatabaseURL     string
	RedisURL        string
}

// Load reads config from env vars, with optional .env file support.
// .env file is loaded if present; missing file is silently ignored.
func Load() *Config {
	_ = godotenv.Load() // no-op if .env absent

	return &Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("APP_ENV", "dev"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ForgeSecretKey:  os.Getenv("FORGE_SECRET_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://forge:forge@localhost:5435/forge?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

// IsDev returns true when running in development mode.
func (c *Config) IsDev() bool {
	return strings.ToLower(c.Env) == "dev"
}

// SlogLevel maps LOG_LEVEL string to a slog.Level.
func (c *Config) SlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
