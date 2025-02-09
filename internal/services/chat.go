package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/pkg/vertex"
	"github.com/yomek33/newln/internal/stores"
)

type ChatService interface {
	CreateChatList(materialID uint) (*models.ChatList, error)
	GetChatListsByMaterialID(materialID uint) ([]models.ChatList, error)
	CreateChat(chatListID uint, userID uuid.UUID, detail string) (*models.Chat, error)
	GetChatsByChatListID(chatListID uint) ([]models.Chat, error)
	CreateMessage(chatID uint, userID uuid.UUID, content string, senderType models.SenderType) (*models.Message, error)
	GetMessagesByChatID(chatID uint) ([]models.Message, error)
	IncrementPendingMessages(chatID uint) error
	ClearPendingMessages(chatID uint) error
	StartChat(chatListID uint, userID uuid.UUID) (*models.Chat, error)
	SendMessage(chatID uint, userID uuid.UUID, message string) (*models.Message, error)
}

type chatService struct {
	store    stores.ChatStore
	vertex   vertex.VertexService
	sessions map[uint]vertex.ChatSession
}

func NewChatService(store stores.ChatStore, vertex vertex.VertexService) ChatService {
	return &chatService{store: store, vertex: vertex}
}

func (s *chatService) CreateChatList(materialID uint) (*models.ChatList, error) {
	//TODO: title
	title := "Chat List"
	if materialID == 0 {
		return nil, errors.New("materialID cannot be zero")
	}
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	chatList := &models.ChatList{
		MaterialID: materialID,
		Title:      title,
	}

	err := s.store.CreateChatList(chatList)
	return chatList, err
}

func (s *chatService) GetChatListsByMaterialID(materialID uint) ([]models.ChatList, error) {
	return s.store.GetChatListsByMaterialID(materialID)
}

func (s *chatService) CreateChat(chatListID uint, userID uuid.UUID, detail string) (*models.Chat, error) {
	if chatListID == 0 {
		return nil, errors.New("chatListID cannot be zero")
	}
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	chat := &models.Chat{
		ChatListID: chatListID,
		Detail:     detail,
	}

	err := s.store.CreateChat(chat)
	return chat, err
}

func (s *chatService) GetChatsByChatListID(chatListID uint) ([]models.Chat, error) {
	return s.store.GetChatsByChatListID(chatListID)
}

func (s *chatService) CreateMessage(chatID uint, userID uuid.UUID, content string, senderType models.SenderType) (*models.Message, error) {
	if chatID == 0 {
		return nil, errors.New("chatID cannot be zero")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	message := &models.Message{
		ChatID:     chatID,
		UserID:     userID,
		Content:    content,
		SenderType: senderType,
	}

	err := s.store.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	err = s.IncrementPendingMessages(chatID)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *chatService) GetMessagesByChatID(chatID uint) ([]models.Message, error) {
	return s.store.GetMessagesByChatID(chatID)
}

func (s *chatService) IncrementPendingMessages(chatID uint) error {
	var chat models.Chat
	chats, err := s.store.GetChatsByChatListID(chatID)
	if err != nil {
		return err
	}
	if len(chats) == 0 {
		return errors.New("chat not found")
	}
	chat = chats[0]

	return s.store.UpdatePendingMessages(chat.ID, chat.PendingMessage+1)
}

func (s *chatService) ClearPendingMessages(chatID uint) error {
	return s.store.UpdatePendingMessages(chatID, 0)
}

// GeminiUser とのチャット生成
func (s *chatService) GenerateSystemMessage(chatID uint, content string) (*models.Message, error) {

	return s.CreateMessage(chatID, models.GeminiUserID, content, models.SenderSystem)
}
