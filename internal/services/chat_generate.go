package services

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/yomek33/newln/internal/models"
)

func (s *chatService) StartChat(chatListID uint, userID uuid.UUID) (*models.Chat, error) {
	if chatListID == 0 {
		return nil, errors.New("chatListID cannot be zero")
	}
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	prompt := "Hello! How can I help you today?"

	chat := &models.Chat{
		ChatListID: chatListID,
	}
	err := s.store.CreateChat(chat)
	if err != nil {
		return nil, err
	}

	session := s.vertex.StartChat(prompt)
	s.sessions[chat.ID] = session

	log.Printf("✅ Chat started: ChatID=%d, UserID=%s", chat.ID, userID)
	return chat, nil
}

func (s *chatService) SendMessage(chatID uint, userID uuid.UUID, message string) (*models.Message, error) {
	if chatID == 0 {
		return nil, errors.New("chatID cannot be zero")
	}
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	// 現在のチャット情報を取得
	chat, err := s.store.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	// メッセージの回数をチェック（10回で終了）
	if chat.PendingMessage >= 10 {
		return &models.Message{
			ChatID:     chatID,
			UserID:     models.GeminiUserID,
			Content:    "🚀 チャット終了！次のステップへ進みます。（後で実装）",
			SenderType: models.SenderSystem,
		}, nil
	}

	// ユーザーのメッセージを DB に保存
	userMessage := &models.Message{
		ChatID:     chatID,
		UserID:     userID,
		Content:    message,
		SenderType: models.SenderUser,
	}
	err = s.store.CreateMessage(userMessage)
	if err != nil {
		return nil, err
	}

	// Gemini にメッセージを送信
	session, exists := s.sessions[chatID]
	if !exists {
		return nil, errors.New("chat session not found")
	}

	ctx := context.Background()
	response, err := session.SendChatMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	// Gemini のレスポンスを DB に保存
	botMessage := &models.Message{
		ChatID:     chatID,
		UserID:     models.GeminiUserID,
		Content:    response,
		SenderType: models.SenderSystem,
	}
	err = s.store.CreateMessage(botMessage)
	if err != nil {
		return nil, err
	}

	// メッセージ回数を更新
	err = s.store.UpdatePendingMessages(chatID, chat.PendingMessage+1)
	if err != nil {
		return nil, err
	}

	log.Printf("💬 User: %s", message)
	log.Printf("🤖 Gemini: %s", response)

	return botMessage, nil
}
