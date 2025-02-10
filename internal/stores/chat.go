package stores

import (
	"errors"

	"github.com/yomek33/newln/internal/models"
	"gorm.io/gorm"
)

type ChatStore interface {
	CreateChatList(chatList *models.ChatList) (*models.ChatList, error)
	GetChatListByMaterialID(materialID uint) (*models.ChatList, error)
	CreateChat(chat *models.Chat) (*models.Chat, error)
	GetChatsByChatListID(chatListID uint) ([]models.Chat, error)
	CreateMessage(message *models.Message) error
	GetMessagesByChatID(chatID uint) ([]models.Message, error)
	UpdatePendingMessages(chatID uint, count uint64) error
	GetChatByID(chatID uint) (*models.Chat, error)
}

type chatStore struct {
	DB *gorm.DB
}

func NewChatStore(db *gorm.DB) ChatStore {
	return &chatStore{DB: db}
}

func (s *chatStore) CreateChatList(chatList *models.ChatList) (*models.ChatList, error) {
    if chatList == nil {
        return nil, errors.New("chatList cannot be nil")
    }
    if chatList.MaterialID == 0 {
        return nil, errors.New("chatList must be linked to a material")
    }
    if chatList.Title == "" {
        return nil, errors.New("chatList title cannot be empty")
    }

    if err := s.DB.Create(chatList).Error; err != nil {
        return nil, err
    }

    if err := s.DB.Preload("Chats").Find(chatList).Error; err != nil {
        return nil, err
    }

    return chatList, nil
}

func (s *chatStore) GetChatListByMaterialID(materialID uint) (*models.ChatList, error) {
	var ChatList models.ChatList

	err := s.DB.Where("material_id = ?", materialID).
		Preload("Chats"). // 関連する Chats も取得
		Find(&ChatList).Error

	return &ChatList, err
}

func (s *chatStore) CreateChat(chat *models.Chat) (*models.Chat, error) {
	if chat == nil {
		return nil, errors.New("chat cannot be nil")
	}
	if chat.ChatListID == 0 {
		return nil, errors.New("chat must be linked to a chat list")
	}

	if err := s.DB.Create(chat).Error; err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *chatStore) GetChatsByChatListID(chatListID uint) ([]models.Chat, error) {
	var chats []models.Chat

	err := s.DB.Where("chat_list_id = ?", chatListID).
		Preload("Messages"). // 関連する Messages も取得
		Find(&chats).Error

	return chats, err
}

func (s *chatStore) CreateMessage(message *models.Message) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	if message.ChatID == 0 {
		return errors.New("message must be linked to a chat")
	}
	if message.Content == "" {
		return errors.New("message content cannot be empty")
	}

	return s.DB.Create(message).Error
}

func (s *chatStore) GetMessagesByChatID(chatID uint) ([]models.Message, error) {
	var messages []models.Message

	err := s.DB.Where("chat_id = ?", chatID).
		Order("created_at ASC"). // 時系列順で取得
		Find(&messages).Error

	return messages, err
}

func (s *chatStore) UpdatePendingMessages(chatID uint, count uint64) error {
	return s.DB.Model(&models.Chat{}).Where("id = ?", chatID).Update("pending_message", count).Error
}

func (s *chatStore) GetChatByID(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	err := s.DB.Where("id = ?", chatID).First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}
