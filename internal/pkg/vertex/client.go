package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"math/rand"

	"cloud.google.com/go/vertexai/genai"
)

const (
	projectID = "newln-448314"
	location  = "us-central1"
	modelName = "gemini-1.5-flash"
)

var semaphore = make(chan struct{}, 3)

type VertexService interface {
	GenerateJsonContent(ctx context.Context, prompt string, jsonSchema *genai.Schema) (json.RawMessage, error)
	StartChat(initialPrompt string) ChatSession
	IsMock() bool
}

func NewVertexService() (VertexService, error) {
	useMock := os.Getenv("USE_MOCK_GEMINI")
	if useMock == "true" {
		fmt.Println("⚡ Using MOCK Vertex Service")
		return NewMockVertexClient(), nil
	}

	fmt.Println("🌍 Using REAL Vertex Service")
	return NewRealVertexClient()
}

type RealVertexClient struct {
	client *genai.Client
}

func NewRealVertexClient() (*RealVertexClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &RealVertexClient{client: client}, nil
}

func (c *RealVertexClient) IsMock() bool {
	return false
}

// 汎用リトライ関数（ジェネリクスを使用）
func retryWithBackoff[T any](_ context.Context, maxRetries int, fn func() (T, error)) (T, error) {
	var err error
	var response T

	for i := 0; i < maxRetries; i++ {
		response, err = fn()
		if err == nil {
			return response, nil
		}

		// 429 (レートリミット) や 500 エラーのときはリトライ
		if isRetryableError(err) {
			waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond // ランダム遅延
			log.Printf("⚠️ API error: %v. Retrying in %v...", err, waitTime+jitter)
			time.Sleep(waitTime + jitter)
		} else {
			return response, err // リトライ不要なエラーは即終了
		}
	}
	return response, fmt.Errorf("API request failed after %d retries: %w", maxRetries, err)
}

// レートリミット (429) や一時的なエラー (500, 504) をチェック
func isRetryableError(err error) bool {
	errStr := err.Error()
	return (errStr == "429 Too Many Requests" || errStr == "500 Internal Server Error" || errStr == "504 Gateway Timeout")
}

func (c *RealVertexClient) GenerateJsonContent(ctx context.Context, prompt string, jsonSchema *genai.Schema) (json.RawMessage, error) {
	log.Printf("🔍 calling GenerateJsonContent: ctx.Err() = %v", ctx.Err())
	model := c.client.GenerativeModel(modelName)
	config := genai.GenerationConfig{
		MaxOutputTokens:  genai.Ptr(int32(8192)),
		TopK:             genai.Ptr(int32(40)),
		TopP:             genai.Ptr(float32(0.95)),
		Temperature:      genai.Ptr(float32(1)),
		ResponseMIMEType: "application/json",
		ResponseSchema:   jsonSchema,
	}
	model.GenerationConfig = config

	// 並列リクエストを制限
	semaphore <- struct{}{}        // スロット確保
	defer func() { <-semaphore }() // スロット解放

	log.Printf("🚀 Sending request to Vertex API with prompt: %s", prompt)

	return retryWithBackoff(ctx, 5, func() (json.RawMessage, error) {
		log.Printf("🔍 Checking ctx.Err() before API call: %v", ctx.Err()) // 追加ログ

		res, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			log.Printf("❌ Failed to generate content: %v", err)
			if ctx.Err() == context.Canceled {
				log.Printf("❌ Context was canceled BEFORE API call: %v", ctx.Err()) // 追加ログ
			}
			return nil, fmt.Errorf("failed to generate content: %w", err)
		}

		if ctx.Err() == context.Canceled {
			log.Printf("❌ Context was canceled AFTER API call: %v", ctx.Err()) // 追加ログ
			return nil, fmt.Errorf("context was canceled after API call")
		}

		log.Printf("✅ Successfully received response from Vertex API")
		return json.RawMessage(res.Candidates[0].Content.Parts[0].(genai.Text)), nil
	})
}

func (c *RealVertexClient) StartChat(initialPrompt string) ChatSession {
	gemini := c.client.GenerativeModel(modelName)
	chat := gemini.StartChat() //  引数なしで StartChat()

	// 初期プロンプトを送信
	session := &realChatSession{chat: chat}
	ctx := context.Background()
	_, err := session.SendChatMessage(ctx, initialPrompt)
	if err != nil {
		log.Printf("⚠️ Failed to send initial prompt: %v", err)
	}

	return session
}

// realChatSession 構造体
type realChatSession struct {
	chat *genai.ChatSession
}

type ChatSession interface {
	SendChatMessage(ctx context.Context, message string) (string, error)
	Close()
}

func (s *realChatSession) SendChatMessage(ctx context.Context, message string) (string, error) {
	log.Printf("🚀 Sending message: %s", message)

	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	response, err := retryWithBackoff[string](ctx, 5, func() (string, error) {
		r, err := s.chat.SendMessage(ctx, genai.Text(message))
		if err != nil {
			return "", err
		}

		responseJSON, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return "", err
		}

		return string(responseJSON), nil
	})

	if err != nil {
		log.Printf("❌ Error sending chat message: %v", err)
		return "", err
	}

	log.Printf("✅ Chat response: %s", response)
	return response, nil
}
func (s *realChatSession) Close() {
	log.Println("🔚 Chat session closed")
}
