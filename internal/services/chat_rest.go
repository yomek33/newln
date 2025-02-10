package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/vertex"
)

type Chatresponse interface{

}

func (s *chatService) CreateFirstChat(chatListID uint) (*models.Chat, []*genai.Content, error) {
	if chatListID == 0 {
		return nil, nil, errors.New("chatListID cannot be zero")
	}

	initalPromptFile , err:= os.ReadFile("./internal/services/prompts/chat.txt")
	if err != nil {
		return nil,nil,  err
	}
	prompt :=  strings.ReplaceAll(string(initalPromptFile), "{{Question}}", "Hello")


	chatRes := vertex.NewSimpleChatSession(s.vertex, prompt)
	chat := &models.Chat{
		ChatListID: chatListID,
		Detail:     "Initial Chat",
	}

	chat, err = s.store.CreateChat(chat)
	if err != nil {
		return nil,nil,  err
	}
	his := chatRes.History

	return chat, his, nil
}
type ChatMessage struct {
	Sender  string `json:"sender"`  // "AI" or "User"
	Message string `json:"message"` // メッセージ内容
}


// **ContinueChat の実装**
func (cs *chatService) ContinueChat(chatID uint, history []ChatMessage, message string) ([]ChatMessage, error) {
	ctx := context.Background()

	// **セッションを取得 or 新規作成**
	session, exists := cs.sessions[chatID]
	if !exists {
		// **CreateFirstChat を呼び出して新規作成**
		_, historyData, err := cs.CreateFirstChat(chatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create first chat session: %w", err)
		}

		initalPromptFile , err:= os.ReadFile("./internal/services/prompts/chat.txt")
		if err != nil {
			return nil,  err
		}
		// **新しいセッションを作成**
		session = vertex.NewSimpleChatSession(cs.vertex, string(initalPromptFile))
		cs.sessions[chatID] = session

		// **履歴を変換**
		updatedHistory := make([]ChatMessage, 0, len(historyData))
		for _, h := range historyData {
			if len(h.Parts) > 0 {
				if text, ok := h.Parts[0].(genai.Text); ok {
					updatedHistory = append(updatedHistory, ChatMessage{
						Sender:  h.Role,
						Message: string(text),
					})
				}
			}
		}

		// **初回の履歴を返す**
		return updatedHistory, nil
	}

	// **ユーザーのメッセージを送信**
	userMessage := &genai.Content{
		Role:  "user",
		Parts: []genai.Part{genai.Text(message)},
	}
	historyData, err := session.SendMessage(ctx, []*genai.Content{userMessage})
	if err != nil {
		return nil, fmt.Errorf("failed to send message to AI: %w", err)
	}

	// **履歴を変換**
	updatedHistory := make([]ChatMessage, 0, len(historyData))
	for _, h := range historyData {
		if len(h.Parts) > 0 {
			if text, ok := h.Parts[0].(genai.Text); ok {
				updatedHistory = append(updatedHistory, ChatMessage{
					Sender:  h.Role,
					Message: string(text),
				})
			}
		}
	}

	return updatedHistory, nil
}
