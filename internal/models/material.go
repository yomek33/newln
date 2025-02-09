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
	UserID               uuid.UUID    `gorm:"type:uuid;not null;index"`
	ULID                 string       `gorm:"type:varchar(255);not null;index;unique"`
	Title                string       `gorm:"type:varchar(255);not null"`
	Content              string       `gorm:"type:text"`
	Status               string       `gorm:"type:material_status;default:'draft'"`
	HasPendingWordList   bool         `gorm:"type:boolean;default:true"`
	HasPendingPhraseList bool         `gorm:"type:boolean;default:true"`
	WordsCount           int          `gorm:"type:int;default:0"`
	PhrasesCount         int          `gorm:"type:int;default:0"`
	WordLists            []WordList   `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
	PhraseLists          []PhraseList `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
	ChatLists            ChatList     `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`
	Summary              string       `gorm:"type:text" json:"-"`
	WordCount            int          `gorm:"type:int;default:0" `
}
