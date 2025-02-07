package services

import (
	"context"
	"fmt"

	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/gemini"
	"github.com/yomek33/newln/internal/stores"
)

type WordService interface {
	GenerateWords(ctx context.Context, materialID uint) ([]models.Word, error)
	GetWordsByMaterialID(materialULID string) ([]models.Word, error)
	GetWordListByMaterialULID(materialULID string) ([]models.WordList, error)
	CreateWordList(wordList *models.WordList) error
	UpdateWordListGenerateStatus(wordListID uint, status string) error
	BulkInsertWords(words []models.Word) error
}

type wordService struct {
	store         stores.WordStore
	materialStore stores.MaterialStore
	geminiClient  gemini.GeminiService
}

func NewWordService(s stores.WordStore, materialStore stores.MaterialStore, geminiClient gemini.GeminiService) WordService {
	return &wordService{store: s, materialStore: materialStore, geminiClient: geminiClient}
}

func (s *wordService) CreateWordList(wordList *models.WordList) error {
	if wordList == nil {
		return fmt.Errorf("wordList cannot be nil")
	}
	return s.store.CreateWordList(wordList)
}

func (s *wordService) GetWordsByMaterialID(materialULID string) ([]models.Word, error) {
	return s.store.GetWordsByMaterialID(materialULID)
}

func (s *wordService) GetWordListByMaterialULID(materialULID string) ([]models.WordList, error) {
	return s.store.GetWordListByMaterialULID(materialULID)
}

func (s *wordService) BulkInsertWords(words []models.Word) error {
	return s.store.BulkInsertWords(words)
}

func (s *wordService) UpdateWordListGenerateStatus(wordListID uint, status string) error {
	return s.store.UpdateWordListGenerateStatus(wordListID, status)
}

func determineWordImportance(_ string) string {
	return "high"
}

func determineWordLevel(_ string) string {
	return "easy"
}
