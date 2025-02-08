package services

import (
	"context"
	"fmt"

	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/vertex"
	"github.com/yomek33/newln/internal/stores"
)

type PhraseService interface {
	GeneratePhrases(ctx context.Context, materialId uint) ([]models.Phrase, error)
	//StorePhrases(materialULID string, phrases []models.Phrase) error
	GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error)
	GetPhraseListByMaterialULID(materialULID string) ([]models.PhraseList, error)
	CreatePhraseList(phraseList *models.PhraseList) error
	BulkInsertPhrases(phrases []models.Phrase) error
	UpdatePhraseListGenerateStatus(phraseListID uint, status string) error
}

type phraseService struct {
	store         stores.PhraseStore
	materialStore stores.MaterialStore
	vertexClient  vertex.VertexService
}

func NewPhraseService(s stores.PhraseStore, materialStore stores.MaterialStore, vertex vertex.VertexService) PhraseService {
	return &phraseService{store: s, materialStore: materialStore, vertexClient: vertex}
}

func (s *phraseService) CreatePhraseList(phraseList *models.PhraseList) error {
	if phraseList == nil {
		return fmt.Errorf("phraseList cannot be nil")
	}

	return s.store.CreatePhraseList(phraseList)
}

func (s *phraseService) GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error) {
	return s.store.GetPhrasesByMaterialID(materialULID)
}

func (s *phraseService) GetPhraseListByMaterialULID(materialULID string) ([]models.PhraseList, error) {
	return s.store.GetPhraseListByMaterialULID(materialULID)
}

func GeneratePhrases(topic string) ([]string, error) {
	return []string{}, nil
}

// func (s *phraseService) GeneratePhrases(ctx context.Context, materialID uint) ([]models.Phrase, error) {
// 	log.Println("Generating phrases")

// 	log.Println("MaterialID", materialID)
// 	material, err := s.materialStore.GetMaterialByID(materialID)
// 	log.Println("Material", material)
// 	if err != nil {
// 		log.Printf("Failed to fetch material: %v", err)
// 		return nil, fmt.Errorf("failed to fetch material: %w", err)
// 	}
// 	if material == nil {
// 		log.Printf("Material is nil")
// 		return nil, fmt.Errorf("material is nil")
// 	}

// 	// Check if VertexClient is nil
// 	// if s.VertexClient == nil {
// 	// 	log.Printf("VertexClient is nil")
// 	// 	return nil, fmt.Errorf("VertexClient is nil")
// 	// }

// 	// log.Printf("Generating phrases for material %d", materialID)

// 	// // Generate phrases using VertexClientx
// 	// phraseTexts, err := s.VertexClient.GeneratePhrases(ctx, material.Content)
// 	// if err != nil {
// 	// 	log.Printf("Failed to generate phrases: %v", err)
// 	// 	return nil, fmt.Errorf("failed to generate phrases: %w", err)
// 	// }
// 	// if phraseTexts == nil {
// 	// 	log.Printf("Generated phrases are nil")
// 	// 	return nil, fmt.Errorf("generated phrases are nil")
// 	// }

// 	// var phrases []models.Phrase
// 	// for _, phraseText := range phraseTexts {
// 	// 	phrases = append(phrases, models.Phrase{
// 	// 		MaterialID: materialID,
// 	// 		Text:       phraseText,
// 	// 		Importance: determineImportance(phraseText),
// 	// 	})
// 	// }

// 	phrases := []models.Phrase{
// 		{
// 			Text:       "phrase1",
// 			Importance: determineImportance("phrase1"),
// 		},
// 		{
// 			Text: "phrase2",

// 			Importance: determineImportance("phrase2"),
// 		},
// 	}
// 	time.Sleep(5 * time.Second)
// 	return phrases, nil
// }

// func (s *phraseService) StorePhrases(materialULID string, phrases []models.Phrase) error {
// 	for _, phrase := range phrases {
// 		if err := s.store.CreatePhrase(&phrase); err != nil {
// 			return fmt.Errorf("failed to store phrase: %w", err)
// 		}
// 	}

// 	return nil
// }

func determineImportance(_ string) string {
	return "high"
}

func (s *phraseService) BulkInsertPhrases(phrases []models.Phrase) error {
	return s.store.BulkInsertPhrases(phrases)
}

func (s *phraseService) UpdatePhraseListGenerateStatus(phraseListID uint, status string) error {
	return s.store.UpdatePhraseListGenerateStatus(phraseListID, status)
}
