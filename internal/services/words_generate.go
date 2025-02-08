package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/gemini"
)

type WordResponse struct {
	ID   int    `json:"id"`
	Word string `json:"word"`
	Pos  string `json:"pos"`
}

type WordWithMeaning struct {
	ID        int    `json:"id"`
	Word      string `json:"word"`
	Pos       string `json:"pos"`
	Meaning   string `json:"meaning"`
	JPMeaning string `json:"jp-meaning"`
}

// `generateWords` を実装
func (s *wordService) GenerateWords(ctx context.Context, materialID uint) ([]models.Word, error) {
	// 1回目のリクエスト（単語と品詞を取得）
	promptFile, err := os.ReadFile("./internal/services/prompts/generate_words.txt")
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
	jsonSchema := gemini.GenerateSchema[[]WordResponse]()
	rawResponse, err := s.geminiClient.GenerateJsonContent(ctx, prompt, jsonSchema)
	if err != nil {
		logger.Error(fmt.Errorf("failed to generate words: %w", err))
		return nil, err
	}

	// JSONをデコード
	wordResponses, err := gemini.DecodeJsonContent[[]WordResponse](rawResponse)
	if err != nil {
		logger.Error(fmt.Errorf("failed to parse JSON: %w", err))
		return nil, err
	}

	logger.Infof("✅ Retrieved %d words: %v", len(wordResponses), wordResponses)

	// 単語リストを5個ずつに分割して並列処理
	chunks := chunkWords(wordResponses, 5)
	logger.Infof("✅ Split words into %d chunks for processing", len(chunks))

	var wg sync.WaitGroup
	resultChan := make(chan []models.Word, len(chunks))
	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(wordsChunk []WordResponse) {
			defer wg.Done()
			logger.Infof("⏳ Generating meaning for %d words", len(wordsChunk))

			// 単語リストを文字列化
			var wordList []string
			for _, word := range wordsChunk {
				wordList = append(wordList, word.Word)
			}
			wordsStr := strings.Join(wordList, ", ")

			meanings, err := s.GenerateWordMeanings(ctx, wordsChunk, wordsStr)
			if err != nil {
				logger.Error(fmt.Errorf("❌ Failed to generate meanings: %w", err))
				errChan <- err
				return
			}

			logger.Infof("✅ Successfully generated meanings for %d words", len(meanings))
			resultChan <- meanings
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	var allWords []models.Word
	finished := false

	for !finished {
		select {
		case words, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				allWords = append(allWords, words...)
			}
		case err, ok := <-errChan:
			if ok {
				return nil, err
			} else {
				errChan = nil
			}
		}
		if resultChan == nil && errChan == nil {
			finished = true
		}
	}

	logger.Infof("🎉 Generated %d words for materialID: %v", len(allWords), materialID)
	return allWords, nil
}

// **単語リストを分割する関数**
func chunkWords(words []WordResponse, chunkSize int) [][]WordResponse {
	var chunks [][]WordResponse
	seen := make(map[string]bool)
	var currentChunk []WordResponse

	for _, word := range words {
		if _, exists := seen[word.Word]; exists {
			continue
		}
		seen[word.Word] = true

		currentChunk = append(currentChunk, word)

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

// **単語の意味を取得する関数**
func (s *wordService) GenerateWordMeanings(ctx context.Context, wordsChunk []WordResponse, wordsStr string) ([]models.Word, error) {
	promptFile, err := os.ReadFile("./internal/services/prompts/generate_words_meanings.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}
	prompt := string(promptFile)
	prompt = strings.ReplaceAll(prompt, "{{TEXT}}", wordsStr)

	jsonSchema := gemini.GenerateSchema[[]WordWithMeaning]()

	rawResponse, err := s.geminiClient.GenerateJsonContent(ctx, prompt, jsonSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate meanings: %w", err)
	}

	meaningResponses, err := gemini.DecodeJsonContent[[]WordWithMeaning](rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var words []models.Word
	for _, res := range meaningResponses {
		words = append(words, models.Word{
			Text:      res.Word,
			Meaning:   res.Meaning,
			JPMeaning: res.JPMeaning,
		})
	}
	logger.Infof("Generated words: %v", words)
	return words, nil
}
