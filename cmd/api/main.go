package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/yomek33/newln/internal/config"
	"github.com/yomek33/newln/internal/handler"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/models/migrations"
	"github.com/yomek33/newln/internal/pkg/vertex"
	"github.com/yomek33/newln/internal/services"
	"github.com/yomek33/newln/internal/stores"

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

	vertexClient, err := vertex.NewVertexService()
	if err != nil {
		log.Fatalf("Failed to create VertexClient: %v", err)
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
	services := services.NewServices(stores, vertexClient)

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
