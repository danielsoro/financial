package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	JWTSecret     string
	Port          string
	StaticDir     string
	AllowedOrigin string
	Tenants       string
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		Port:          os.Getenv("PORT"),
		StaticDir:     os.Getenv("STATIC_DIR"),
		AllowedOrigin: os.Getenv("ALLOWED_ORIGIN"),
		Tenants:       os.Getenv("TENANTS"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}
