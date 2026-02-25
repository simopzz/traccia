package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
	Environment   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Debug(".env file not found, using system environment variables")
	}

	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":3000"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable"),
		Environment:   getEnv("ENVIRONMENT", "development"),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
