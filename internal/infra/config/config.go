package config

import (
	"os"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
	Environment   string
}

func Load() *Config {
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
