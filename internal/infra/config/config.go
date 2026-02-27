package config

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":3000"`
	DatabaseURL   string `env:"DATABASE_URL" envDefault:"postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable"`
	Environment   string `env:"ENVIRONMENT" envDefault:"development"`
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Debug(".env file not found, using system environment variables")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse environment variables", "error", err)
		panic(err)
	}

	return &cfg
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
