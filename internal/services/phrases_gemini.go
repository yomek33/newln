package services

import (
	"context"
	"fmt"

	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/gemini"
)

// Geminiからのレスポンス
type PhraseResponse struct {
	ID          int    `json:"id"`
	Collocation string `json:"collocation"`
	FromText    bool   `json:"from_text"`
	Example     string `json:"example"`
	Difficulty  string `json:"difficulty"`
}

func (s *phraseService) GeneratePhrases(ctx context.Context, materialID uint) ([]models.Phrase, error) {
	material, err := s.materialStore.GetMaterialByID(materialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get material: %w", err)
	}

	rawResponse, err := s.geminiClient.GenerateJsonContent(ctx, material.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate phrases: %w", err)
	}
	response, err := gemini.DecodeJsonContent[PhraseResponse](rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var phrases []models.Phrase
	for _, res := range response {
		phrases = append(phrases, models.Phrase{
			PhraseListID: materialID,
			Text:         res.Collocation,
			Importance:   determineImportance(res.Difficulty),
		})
	}

	logger.Info(fmt.Sprintf("Generated phrases: %+v", phrases))

	return phrases, nil
}
