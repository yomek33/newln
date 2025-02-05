package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/genai"
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

func (c *RealGeminiClient) GenerateJsonContent(ctx context.Context, prompt string) (json.RawMessage, error) {
	model := "gemini-1.5-flash"

	config := genai.GenerateContentConfig{
		MaxOutputTokens:  genai.Ptr(int64(8192)),
		TopK:            genai.Ptr(float64(40)),
		TopP:            genai.Ptr(0.95),
		Temperature:     genai.Ptr(float64(1)),
		ResponseMIMEType: "application/json",
	}

	res, err := c.client.Models.GenerateContent(ctx, model, genai.Text(prompt), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	partStr := res.Candidates[0].Content.Parts[0].Text

	return json.RawMessage(partStr), nil
}