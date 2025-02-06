package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"math/rand"

	"google.golang.org/genai"
)

var semaphore = make(chan struct{}, 5)

type GeminiService interface {
	GenerateJsonContent(ctx context.Context, prompt string, jsonSchema *genai.Schema) (json.RawMessage, error)
	IsMock() bool
}

func NewGeminiService(ctx context.Context, apiKey string) (GeminiService, error) {
	useMock := os.Getenv("USE_MOCK_GEMINI")

	if useMock == "true" {
		fmt.Println("⚡ Using MOCK Gemini Service")
		return NewMockGeminiClient(), nil
	}

	fmt.Println("🌍 Using REAL Gemini Service")
	return NewRealGeminiClient(ctx, apiKey)
}

type RealGeminiClient struct {
	client *genai.Client
}

func NewRealGeminiClient(ctx context.Context, apiKey string) (*RealGeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &RealGeminiClient{client: client}, nil
}

func (c *RealGeminiClient) IsMock() bool {
	return false
}

func retryWithBackoff(ctx context.Context, maxRetries int, fn func() (json.RawMessage, error)) (json.RawMessage, error) {
	var err error
	var response json.RawMessage

	for i := 0; i < maxRetries; i++ {
		response, err = fn()
		if err == nil {
			return response, nil // 成功
		}

		// 429 (レートリミット) や 500 エラーのときはリトライ
		if isRetryableError(err) {
			waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond // ランダム遅延
			log.Printf("⚠️ API error: %v. Retrying in %v...", err, waitTime+jitter)
			time.Sleep(waitTime + jitter)
		} else {
			return nil, err // リトライ不要なエラーは即終了
		}
	}
	return nil, fmt.Errorf("API request failed after %d retries: %w", maxRetries, err)
}

// レートリミット (429) や一時的なエラー (500, 504) をチェック
func isRetryableError(err error) bool {
	errStr := err.Error()
	return (errStr == "429 Too Many Requests" || errStr == "500 Internal Server Error" || errStr == "504 Gateway Timeout")
}


func (c *RealGeminiClient) GenerateJsonContent(ctx context.Context, prompt string, jsonSchema *genai.Schema) (json.RawMessage, error) {
	model := "gemini-1.5-flash"
	config := genai.GenerateContentConfig{
		MaxOutputTokens:  genai.Ptr(int64(8192)),
		TopK:             genai.Ptr(float64(40)),
		TopP:             genai.Ptr(0.95),
		Temperature:      genai.Ptr(float64(1)),
		ResponseMIMEType: "application/json",
		ResponseSchema: 	   jsonSchema,
	}

	// セマフォを使って並列リクエストを制限
	semaphore <- struct{}{} // スロット確保
	defer func() { <-semaphore }() // スロット解放

	return retryWithBackoff(ctx, 3, func() (json.RawMessage, error) {
		log.Printf("🚀 Sending request to Gemini API (model: %s)", model)

		// APIリクエスト
		res, err := c.client.Models.GenerateContent(ctx, model, genai.Text(prompt), &config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate content: %w", err)
		}

		//レスポンスが空でないかチェック
		if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
			return nil, fmt.Errorf("no content generated")
		}

		part := res.Candidates[0].Content.Parts[0]
		if part.Text == "" {
			return nil, fmt.Errorf("response text is empty")
		}

		log.Printf("✅ Successfully received response from Gemini API")
		return json.RawMessage(part.Text), nil
	})
}