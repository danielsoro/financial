package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	StaticDir      string
	AllowedOrigin  string
	AppURL         string
	SendGridAPIKey string
	EmailFrom      string
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		Port:           os.Getenv("PORT"),
		StaticDir:      os.Getenv("STATIC_DIR"),
		AllowedOrigin:  os.Getenv("ALLOWED_ORIGIN"),
		AppURL:         os.Getenv("APP_URL"),
		SendGridAPIKey: os.Getenv("SENDGRID_API_KEY"),
		EmailFrom:      os.Getenv("EMAIL_FROM"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.AppURL == "" {
		cfg.AppURL = "http://localhost:5173"
	}

	return cfg
}
