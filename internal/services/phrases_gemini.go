package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/vertex"
)

// Vertexからのレスポンス
type PhraseResponse struct {
	Phrase     string `json:"phrase"`
	FromText   bool   `json:"from_text"`
	Example    string `json:"example"`
	Difficulty string `json:"difficulty"`
}

func (s *phraseService) GeneratePhrases(ctx context.Context, materialID uint) ([]models.Phrase, error) {
	logger.Infof("🚀 Start GeneratePhrases for materialID: %v", materialID)
	promptFile, err := os.ReadFile("./internal/services/prompts/generate_phrases.txt")
	if err != nil {
		logger.Error(fmt.Errorf("failed to read prompt file: %w", err))
		return nil, err
	}
	prompt := string(promptFile)

	material, err := s.materialStore.GetMaterialByID(materialID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get material: %w", err))
		return nil, err
	}

	prompt = strings.ReplaceAll(prompt, "{{TEXT}}", material.Content)

	jsonSchema := vertex.GenerateSchema[[]PhraseResponse]()
	rawResponse, err := s.vertexClient.GenerateJsonContent(ctx, prompt, jsonSchema)
	if err != nil {
		logger.Error(fmt.Errorf("failed to generate phrases: %w", err))
		return nil, err
	}

	// JSONリストを正しくデコード
	phraseResponses, err := vertex.DecodeJsonContent[[]PhraseResponse](rawResponse)
	if err != nil {
		logger.Error(fmt.Errorf("failed to parse JSON: %w", err))
		return nil, err
	}

	logger.Infof("✅ Genrerated Phrases: %v", phraseResponses)

	// quotaが小さすぎる。。。
	chunks := chunkAndDeduplicatePhrases(phraseResponses, 30)
	logger.Infof("✅ Split phrases into %d chunks for processing", len(chunks))

	// 並列処理で意味を生成
	var wg sync.WaitGroup
	resultChan := make(chan []models.Phrase, len(chunks))
	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(phrasesChunk []PhraseResponse) {
			defer wg.Done()
			logger.Infof("⏳ Generating meaning for %d phrases", len(phrasesChunk))

			var phraseList []string
			for _, phrase := range phrasesChunk {
				phraseList = append(phraseList, phrase.Phrase)
			}
			phrasesStr := strings.Join(phraseList, ", ")

			meanings, err := s.GenerateMeaning(ctx, phrasesStr)
			if err != nil {
				logger.Error(fmt.Errorf("❌ Failed to generate meaning: %w", err))
				errChan <- err
				return
			}

			logger.Infof("✅ Successfully generated meanings for %d phrases", len(meanings))
			resultChan <- meanings
		}(chunk)
	}

	go func() {
		wg.Wait()
		logger.Infof("✅ All phrase meaning generation completed, closing channels")
		close(resultChan)
		close(errChan)
	}()

	var allPhrases []models.Phrase
	finished := false

	for !finished {
		select {
		case phrases, ok := <-resultChan:
			if !ok {
				logger.Infof("⚠️ resultChan closed")
				resultChan = nil
			} else {
				logger.Infof("📦 Received %d phrases from resultChan", len(phrases))
				allPhrases = append(allPhrases, phrases...)
			}
		case err, ok := <-errChan:
			if ok {
				logger.Error(fmt.Errorf("❌ Error received from errChan: %w", err))
				return nil, err
			} else {
				logger.Infof("⚠️ errChan closed")
				errChan = nil
			}
		}
		if resultChan == nil && errChan == nil {
			finished = true
		}
	}
	logger.Infof("🎉 Generated %d phrases for materialID: %v", len(allPhrases), materialID)
	return allPhrases, nil
}

func chunkAndDeduplicatePhrases(phrases []PhraseResponse, chunkSize int) [][]PhraseResponse {
	var chunks [][]PhraseResponse
	seen := make(map[string]bool)
	var currentChunk []PhraseResponse

	for _, phrase := range phrases {
		if _, exists := seen[phrase.Phrase]; exists {
			continue
		}
		seen[phrase.Phrase] = true

		currentChunk = append(currentChunk, phrase)

		if len(currentChunk) == chunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = nil
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

type PhraseWithMeaning struct {
	Phrase     string `json:"phrase"`
	FromText   bool   `json:"from_text"`
	Example    string `json:"example"`
	Difficulty string `json:"difficulty"`
	JPMeaning  string `json:"jp_meaning"`
	Meaning    string `json:"meaning"`
}

// 意味を生成する関数
func (s *phraseService) GenerateMeaning(ctx context.Context, phrasesStr string) ([]models.Phrase, error) {
	promptFile, err := os.ReadFile("./internal/services/prompts/generate_meanings_phrases.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}
	prompt := string(promptFile)
	prompt = strings.ReplaceAll(prompt, "{{INPUT}}", phrasesStr)

	jsonSchema := vertex.GenerateSchema[[]PhraseWithMeaning]()

	rawResponse, err := s.vertexClient.GenerateJsonContent(ctx, prompt, jsonSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate meanings: %w", err)
	}

	meaningResponses, err := vertex.DecodeJsonContent[[]PhraseWithMeaning](rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var phrases []models.Phrase
	for _, res := range meaningResponses {
		phrases = append(phrases, models.Phrase{
			Text:       res.Phrase,
			Meaning:    res.Meaning,
			JPMeaning:  res.JPMeaning,
			Example:    res.Example,
			FromText:   res.FromText,
			Difficulty: res.Difficulty,
			Importance: determineImportance(res.Difficulty),
		})
	}

	return phrases, nil
}
