package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	Detail         string `gorm:"type:text"`
	ChatListID     uint   `gorm:"not null;index"`
	PendingMessage uint64	`gorm:"type:bigint;default:0" json:"-"`
	Messages       []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
}

type Message struct {
	gorm.Model
	ChatID     uint   `gorm:"not null;index"`
	Content    string `gorm:"type:text"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	SenderType SenderType `gorm:"type:ENUM('user', 'system');default:'user'"`
}
type SenderType string

const (
	SenderUser   SenderType = "user"
	SenderSystem SenderType = "system"
)
var GeminiUserID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
// チャットリストのデータ構造
type ChatList struct {
	gorm.Model
	MaterialID         uint   `gorm:"not null;index"`
	Title              string `gorm:"type:varchar(255);not null"`
	Chats              []Chat `gorm:"foreignKey:ChatListID;constraint:OnDelete:CASCADE"`
	SuggestedQuestions JSONStringArray `gorm:"type:text" json:"-"` 
}


type JSONStringArray []string

func (j *JSONStringArray) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONStringArray) Value() (interface{}, error) {
	return json.Marshal(j)
}