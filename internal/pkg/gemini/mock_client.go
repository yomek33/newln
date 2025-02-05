package gemini

import (
	"context"
	"encoding/json"
	"fmt"
)

// MockGeminiClient（開発 & テスト用）
type MockGeminiClient struct{}

func NewMockGeminiClient() *MockGeminiClient {
	return &MockGeminiClient{}
}

func (m *MockGeminiClient) GenerateJsonContent(ctx context.Context, prompt string) (json.RawMessage, error) {
    fmt.Println("⚡ Using MOCK Gemini Service")

    // 空の JSON データを返す
    return json.RawMessage(`[]`), nil
}

func (m *MockGeminiClient) IsMock() bool {
	return true
}
