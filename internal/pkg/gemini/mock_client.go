package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

// MockGeminiClient（テスト用）
type MockGeminiClient struct {
	ResponseData json.RawMessage
}

func NewMockGeminiClient(responseData ...json.RawMessage) *MockGeminiClient {
	defaultResponse := json.RawMessage(`[]`)
	if len(responseData) > 0 {
		defaultResponse = responseData[0]
	}

	return &MockGeminiClient{
		ResponseData: defaultResponse,
	}
}

func (m *MockGeminiClient) GenerateJsonContent(ctx context.Context, prompt string, jsonSchema *genai.Schema) (json.RawMessage, error) {
	fmt.Println("⚡ Using MOCK Gemini Service")

	// 設定されたモックレスポンスを返す
	if len(m.ResponseData) == 0 {
		return json.RawMessage(`[]`), nil
	}
	return m.ResponseData, nil
}

func (m *MockGeminiClient) IsMock() bool {
	return true
}
