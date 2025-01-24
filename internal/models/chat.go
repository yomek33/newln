package models

import "gorm.io/gorm"

type Chat struct {
	gorm.Model
	Detail         string `gorm:"type:text"`
	MaterialID     uint   `gorm:"not null;index"`
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
