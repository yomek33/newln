package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
    StatusDraft     = "draft"
    StatusPublished = "published"
    StatusArchived  = "archived"
)

type Material struct {
    gorm.Model
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    Title     string    `gorm:"type:varchar(255);not null"`
    Content   string    `gorm:"type:text"`
    Status    string    `gorm:"type:material_status;default:'draft'"`
    Words     []Word    `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
    Phrases   []Phrase  `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
    Chats     []Chat    `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
}