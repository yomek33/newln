package services

import (
	"context"
	"fmt"
	"log"

	"newln/internal/models"
	"newln/internal/sse"
	"newln/internal/stores"
)

type WordService interface {
	GenerateWords(ctx context.Context, materialID uint) ([]models.Word, error)
	GetWordsByMaterialID(materialULID string) ([]models.Word, error)
	GetWordListByMaterialULID(materialULID string) ([]models.WordList, error)
	CreateWordList(wordList *models.WordList) error
	UpdateWordListGenerateStatus(wordListID uint, status string) error
	BulkInsertWords(words []models.Word) error
	HandleWordGeneration(ctx context.Context, materialID uint, sseManager *sse.SSEManager)error
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

func (s *wordService) HandleWordGeneration(ctx context.Context, materialID uint, sseManager *sse.SSEManager) error {
    wordList := models.WordList{
        MaterialID:    materialID,
        Title:         "Default Word List",
        GenerateStatus: "pending",
    }

    if err := s.CreateWordList(&wordList); err != nil {
        return fmt.Errorf("failed to create word list: %w", err)
    }
    s.UpdateWordListGenerateStatus(wordList.ID, "processing")
    sseManager.Broadcast(fmt.Sprintf(`{"word_list_id":%d, "status":"processing"}`, wordList.ID))

    words, err := s.GenerateWords(ctx, materialID)
    if err != nil {
        s.UpdateWordListGenerateStatus(wordList.ID, "failed")
        sseManager.Broadcast(fmt.Sprintf(`{"word_list_id":%d, "status":"failed"}`, wordList.ID))
        return fmt.Errorf("failed to generate words: %w", err)
    }
    if len(words) == 0 {
        s.UpdateWordListGenerateStatus(wordList.ID, "failed")
        sseManager.Broadcast(fmt.Sprintf(`{"word_list_id":%d, "status":"failed"}`, wordList.ID))
        return fmt.Errorf("no words generated")
    }

    for i := range words {
        words[i].WordListID = wordList.ID
    }
    if err := s.BulkInsertWords(words); err != nil {
        s.UpdateWordListGenerateStatus(wordList.ID, "failed")
        sseManager.Broadcast(fmt.Sprintf(`{"word_list_id":%d, "status":"failed"}`, wordList.ID))
        return fmt.Errorf("failed to store words: %w", err)
    }

    s.UpdateWordListGenerateStatus(wordList.ID, "completed")
    sseManager.Broadcast(fmt.Sprintf(`{"word_list_id":%d, "status":"completed"}`, wordList.ID))
    return nil
}