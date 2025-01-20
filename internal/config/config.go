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
}

const (
	// expires cookie expiration time
	SessionDuration = time.Hour * 10
)

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{
		UseSSL:       os.Getenv("USE_SSL"),
		Port:         os.Getenv("PORT"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		SupabaseURI:  os.Getenv("SUPABASE_URI"),
	}

	if cfg.Port == "" || cfg.UseSSL == "" || cfg.GeminiAPIKey == "" || cfg.SupabaseURI == "" {
		return nil, fmt.Errorf("one or more required environment variables are missing")
	}

	return cfg, nil
}
