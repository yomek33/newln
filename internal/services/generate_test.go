package services

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/gemini"
	stores_mock "github.com/yomek33/newln/internal/stores/mocks"
	"gorm.io/gorm"
)

func TestPhraseService_GeneratePhrases_RealGemini(t *testing.T) {
	if err := godotenv.Load("./../../.env"); err != nil {
		t.Fatalf("error loading .env file: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Fatal("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()

	geminiClient, err := gemini.NewRealGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("failed to create Gemini client: %v", err)
	}

	materialStore := stores_mock.NewMockMaterialStore()
	materialStore.Materials[1] = &models.Material{
		Model:   gorm.Model{ID: 1},
		Content: "hello",
	}

	service := &phraseService{
		materialStore: materialStore,
		geminiClient:  geminiClient,
	}

	materialID := uint(1)

	// テスト実行
	phrases, err := service.GeneratePhrases(ctx, materialID)
	if err != nil {
		t.Fatalf("GeneratePhrases failed: %v", err)
	}

	// 期待するデータがあるかチェック
	if len(phrases) == 0 {
		t.Fatalf("Expected phrases, but got empty response")
	}

	t.Logf("Success! Generated phrases: %+v", phrases)
}

func TestPhraseService_GeneratePhrases_MockGemini(t *testing.T) {
	ctx := context.Background()

	geminiClient := gemini.NewMockGeminiClient()
	materialStore := stores_mock.NewMockMaterialStore()
	materialStore.Materials[1] = &models.Material{
		Model:   gorm.Model{ID: 1},
		Content: "hello",
	}

	service := &phraseService{
		materialStore: materialStore,
		geminiClient:  geminiClient,
	}

	materialID := uint(1)

	// テスト実行
	phrases, err := service.GeneratePhrases(ctx, materialID)
	if err != nil {
		t.Fatalf("GeneratePhrases failed: %v", err)
	}

	// 期待するデータがあるかチェック
	if len(phrases) == 0 {
		t.Fatalf("Expected phrases, but got empty response")
	}

	t.Logf("Success! Generated phrases: %+v", phrases)
}
