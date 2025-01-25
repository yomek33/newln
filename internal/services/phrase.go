package services

import (
	"context"
	"fmt"
	"log"

	"newln/internal/models"
	"newln/internal/stores"

	"github.com/google/uuid"
)

type PhraseService interface {
	GeneratePhrases(ctx context.Context, materialID uint, UserID uuid.UUID) ([]models.Phrase, error)
	StorePhrases(materialID uint, phrases []models.Phrase) error
	GetPhrasesByMaterialID(materialID uint) ([]models.Phrase, error)
}

type phraseService struct {
	store           stores.PhraseStore
	materialStore stores.MaterialStore
	// GeminiClient    *gemini.Client
}

func NewPhraseService(s stores.PhraseStore, materialStore stores.MaterialStore) PhraseService {
	return &phraseService{store: s, materialStore: materialStore}
}

func (s *phraseService) StorePhrase(phrase *models.Phrase) error {
	return s.store.CreatePhrase(phrase)
}

func (s *phraseService) GetPhrasesByMaterialID(materialID uint) ([]models.Phrase, error) {
	return s.store.GetPhrasesByMaterialID(materialID)
}

func GeneratePhrases(topic string) ([]string, error) {
	return []string{}, nil
}

func (s *phraseService) GeneratePhrases(ctx context.Context, materialID uint, UserID uuid.UUID) ([]models.Phrase, error) {
	log.Println("Generating phrases")

	log.Println("MaterialID", materialID)
	log.Println("UserID", UserID)
	material, err := s.materialStore.GetMaterialByID(materialID, UserID)
	log.Println("Material", material)
	if err != nil {
		log.Printf("Failed to fetch material: %v", err)
		return nil, fmt.Errorf("failed to fetch material: %w", err)
	}
	if material == nil {
		log.Printf("Material is nil")
		return nil, fmt.Errorf("material is nil")
	}

	// Check if GeminiClient is nil
	// if s.GeminiClient == nil {
	// 	log.Printf("GeminiClient is nil")
	// 	return nil, fmt.Errorf("GeminiClient is nil")
	// }

	// log.Printf("Generating phrases for material %d", materialID)

	// // Generate phrases using GeminiClientx
	// phraseTexts, err := s.GeminiClient.GeneratePhrases(ctx, material.Content)
	// if err != nil {
	// 	log.Printf("Failed to generate phrases: %v", err)
	// 	return nil, fmt.Errorf("failed to generate phrases: %w", err)
	// }
	// if phraseTexts == nil {
	// 	log.Printf("Generated phrases are nil")
	// 	return nil, fmt.Errorf("generated phrases are nil")
	// }

	// var phrases []models.Phrase
	// for _, phraseText := range phraseTexts {
	// 	phrases = append(phrases, models.Phrase{
	// 		MaterialID: materialID,
	// 		Text:       phraseText,
	// 		Importance: determineImportance(phraseText),
	// 	})
	// }

	phrases := []models.Phrase{
		{
			MaterialID: materialID,
			Text:       "phrase1",
			Importance: determineImportance("phrase1"),
		},
		{
			MaterialID: materialID,
			Text:       "phrase2",

			Importance: determineImportance("phrase2"),
		},
	}
	return phrases, nil
}

func (s *phraseService) StorePhrases(materialID uint, phrases []models.Phrase) error {
	for _, phrase := range phrases {
		if err := s.store.CreatePhrase(&phrase); err != nil {
			return fmt.Errorf("failed to store phrase: %w", err)
		}
	}

	return nil
}

func determineImportance(_ string) string {
	return "high"
}
