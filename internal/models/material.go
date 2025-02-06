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
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	ULID        string       `gorm:"type:varchar(255);not null;index;unique"`
	Title       string       `gorm:"type:varchar(255);not null"`
	Content     string       `gorm:"type:text"`
	Status      string       `gorm:"type:material_status;default:'draft'"`
	WordLists   []WordList   `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
	PhraseLists []PhraseList `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
	ChatLists   []ChatList   `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
}
