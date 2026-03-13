package config

import (
	"log/slog"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/zenv"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string `zog:"SERVER_ADDRESS"`
	DatabaseURL   string `zog:"DATABASE_URL"`
	Environment   string `zog:"ENVIRONMENT"`
}

var configSchema = z.Struct(z.Shape{
	"ServerAddress": z.String().Default(":3000"),
	"DatabaseURL":   z.String().Default("postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable"),
	"Environment":   z.String().Default("development"),
})

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Debug(".env file not found, using system environment variables")
	}

	var cfg Config
	errs := configSchema.Parse(zenv.NewDataProvider(), &cfg)
	if errs != nil {
		slog.Error("Failed to parse environment variables", "error", errs)
		panic(errs)
	}

	return &cfg
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
