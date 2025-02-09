package vertex

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
)

// MockVertexClient（テスト用）
type MockVertexClient struct {
	ResponseData json.RawMessage
}

func NewMockVertexClient(responseData ...json.RawMessage) *MockVertexClient {
	defaultResponse := json.RawMessage(`[]`)
	if len(responseData) > 0 {
		defaultResponse = responseData[0]
	}

	return &MockVertexClient{
		ResponseData: defaultResponse,
	}
}

func (m *MockVertexClient) GenerateJsonContent(ctx context.Context, prompt string, jsonSchema *genai.Schema) (json.RawMessage, error) {
	fmt.Println("⚡ Using MOCK Vertex Service")

	if len(m.ResponseData) == 0 {
		return json.RawMessage(`[]`), nil
	}
	return m.ResponseData, nil
}

func (m *MockVertexClient) IsMock() bool {
	return true
}

func (m *MockVertexClient) StartChat(initialPrompt string) ChatSession {
    return MockChatSession{}
}

type MockChatSession struct{}

func (m MockChatSession) SendChatMessage(ctx context.Context, message string) (string, error) {
	return fmt.Sprintf("MOCK RESPONSE: %s", message), nil
}

func (m MockChatSession) Close() {
	fmt.Println("🔚 Closing Mock Chat Session")
}
