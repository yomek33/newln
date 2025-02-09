package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	Detail         string `gorm:"type:text"`
	ChatListID     uint   `gorm:"not null;index"`
	UserID         string `gorm:"type:varchar(255);not null;index"`
	PendingMessage uint64
	Messages       []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
}

type Message struct {
	gorm.Model
	ChatID     uint   `gorm:"not null;index"`
	Content    string `gorm:"type:text"`
	UserID     string `gorm:"type:varchar(255)"`
	SenderType string `gorm:"type:sender_type;default:'user'"`
}

// チャットリストのデータ構造
type ChatList struct {
	gorm.Model
	MaterialID         uint   `gorm:"not null;index"`
	Title              string `gorm:"type:varchar(255);not null"`
	Chats              []Chat `gorm:"foreignKey:ChatListID;constraint:OnDelete:CASCADE"`
	SuggestedQuestions JSONStringArray `gorm:"type:text"` 
}


type JSONStringArray []string

func (j *JSONStringArray) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONStringArray) Value() (interface{}, error) {
	return json.Marshal(j)
}