package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService interface {
	GenerateJsonContent(ctx context.Context, prompt string) (json.RawMessage, error)
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &RealGeminiClient{client: client}, nil
}

func (c *RealGeminiClient) Close() {
	c.client.Close()
}

func (c *RealGeminiClient) IsMock() bool {
	return false
}

func (c *RealGeminiClient) GenerateJsonContent(ctx context.Context, prompt string) (json.RawMessage, error) {
	model := c.client.GenerativeModel("gemini-1.5-flash")
	model.ResponseMIMEType = "application/json"

	res, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	partStr, ok := res.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("response is not a valid string")
	}

	return json.RawMessage(partStr), nil
}

func DecodeJsonContent[T any](data json.RawMessage) ([]T, error) {
	var output []T
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return output, nil
}
