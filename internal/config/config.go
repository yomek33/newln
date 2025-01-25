package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	UseSSL       string
	Port         string
	GeminiAPIKey string
	SupabaseURI  string
	JwtSecret    []byte
}

const (
	SessionDuration = time.Hour * 10
)

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{
		UseSSL:       os.Getenv("USE_SSL"),
		Port:         os.Getenv("PORT"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		SupabaseURI:  os.Getenv("SUPABASE_URI"),
		JwtSecret:    []byte(os.Getenv("JWT_SECRET_KEY")),
	}
	fmt.Println(os.Getenv("JWT_SECRET_KEY"))

	if cfg.Port == "" || cfg.UseSSL == "" || cfg.GeminiAPIKey == "" || cfg.SupabaseURI == "" || cfg.JwtSecret == nil {
		return nil, fmt.Errorf("one or more required environment variables are missing")
	}

	return cfg, nil
}
