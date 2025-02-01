package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"newln/internal/models"
	"newln/internal/stores"
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
}

func NewWordService(s stores.WordStore, materialStore stores.MaterialStore) WordService {
	return &wordService{store: s, materialStore: materialStore}
}

// ✅ `WordList` を作成
func (s *wordService) CreateWordList(wordList *models.WordList) error {
	if wordList == nil {
		return fmt.Errorf("wordList cannot be nil")
	}
	return s.store.CreateWordList(wordList)
}

// ✅ `MaterialID` から `Words` を取得
func (s *wordService) GetWordsByMaterialID(materialULID string) ([]models.Word, error) {
	return s.store.GetWordsByMaterialID(materialULID)
}

// ✅ `MaterialID` から `WordList` を取得
func (s *wordService) GetWordListByMaterialULID(materialULID string) ([]models.WordList, error) {
	return s.store.GetWordListByMaterialULID(materialULID)
}

// ✅ `Words` を一括挿入
func (s *wordService) BulkInsertWords(words []models.Word) error {
	return s.store.BulkInsertWords(words)
}

// ✅ `Words` を生成（ダミーデータとして2つ）
func (s *wordService) GenerateWords(ctx context.Context, materialID uint) ([]models.Word, error) {
	log.Println("Generating words")

	log.Println("MaterialID", materialID)
	material, err := s.materialStore.GetMaterialByID(materialID)
	log.Println("Material", material)
	if err != nil {
		log.Printf("Failed to fetch material: %v", err)
		return nil, fmt.Errorf("failed to fetch material: %w", err)
	}
	if material == nil {
		log.Printf("Material is nil")
		return nil, fmt.Errorf("material is nil")
	}

	words := []models.Word{
		{
			Text:       "word1",
			Importance: determineWordImportance("word1"),
			Level:      determineWordLevel("word1"),
			Meaning:    "意味1",
		},
		{
			Text:       "word2",
			Importance: determineWordImportance("word2"),
			Level:      determineWordLevel("word2"),
			Meaning:    "意味2",
		},
	}

	time.Sleep(5 * time.Second)
	return words, nil
}

func (s *wordService) UpdateWordListGenerateStatus(wordListID uint, status string) error {
	return s.store.UpdateWordListGenerateStatus(wordListID, status)
}

// ✅ 重要度を決定（ダミー）
func determineWordImportance(_ string) string {
	return "high"
}

// ✅ レベルを決定（ダミー）
func determineWordLevel(_ string) string {
	return "beginner"
}
