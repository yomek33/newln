package models

import (
	"gorm.io/gorm"
)

type Phrase struct {
	gorm.Model
	PhraseListID uint   `gorm:"not null;index"`
<<<<<<< HEAD
	Text       string `gorm:"type:varchar(255);not null"`
	Importance string `gorm:"type:importance_level;default:'medium'"`
=======
	Text         string `gorm:"type:varchar(255);not null"`
	Importance   string `gorm:"type:importance_level;default:'medium'"`
}
type PhraseList struct {
	gorm.Model
	MaterialID     uint     `gorm:"type:uint;not null;index"`
	Title          string   `gorm:"type:varchar(255);not null"`
	Phrases        []Phrase `gorm:"foreignKey:PhraseListID;constraint:OnDelete:CASCADE"`
	GenerateStatus string   `gorm:"type:phrase_list_status;default:'pending'"`
>>>>>>> 438e269 (feat(handler): integrate WordService and SSEManager into Handlers for enhanced functionality)
}
type PhraseList struct {
    gorm.Model
    MaterialID uint `gorm:"type:uint;not null;index"`
    Title      string    `gorm:"type:varchar(255);not null"`
    Phrases    []Phrase  `gorm:"foreignKey:PhraseListID;constraint:OnDelete:CASCADE"`
	GenerateStatus string `gorm:"type:phrase_list_status;default:'pending'"`
}
