package main

import (
	"context"
	"fmt"
	"log"
	"newln/internal/config"
	"newln/internal/handler"
	"newln/internal/models"
	"newln/internal/models/migrations"
	"newln/internal/pkg/gemini"
	"newln/internal/services"
	"newln/internal/stores"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type application struct {
	DB *gorm.DB
}

func main() {
	// Initialize Echo
	e := handler.Echo()
	e.Validator = handler.NewValidator()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Gemini クライアントの作成 (Mock or Real)
	geminiClient, err := gemini.NewGeminiService(context.Background(), cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini service: %v", err)
	}

	// Build DSN
	dsn := cfg.SupabaseURI
	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize application structure
	app := &application{DB: db}

	stores := stores.NewStores(app.DB)
	services := services.NewServices(stores, geminiClient)

	h := handler.NewHandler(services, cfg.JwtSecret)

	// ルート設定
	h.SetDefault(e)
	h.SetAPIRoutes(e)

	// DBマイグレーション
	if err := migrations.CreateEnumTypes(db); err != nil {
		log.Fatalf("failed to create enums: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Material{},
		&models.Word{},
		&models.Phrase{},
		&models.Progress{},
		&models.Chat{},
		&models.Message{},
	); err != nil {
		log.Fatalf("failed to run auto-migration: %v", err)
	}

	// サーバー起動
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
