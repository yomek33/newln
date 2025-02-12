package models

import (
	"gorm.io/gorm"
)

const (
	Advance     = "advance"
	Intemediate = "intermediate"
	Easy        = "easy"
)

type Phrase struct {
	gorm.Model
	PhraseListID uint   `gorm:"not null;index"`
	Text         string `gorm:"type:varchar(255);not null"`
	Importance   string `gorm:"type:importance_level;default:'medium'"`
	Meaning      string `gorm:"type:text"`
	JPMeaning    string `gorm:"type:text"`
	Example      string `gorm:"type:text"`
	FromText     bool   `gorm:"type:boolean;default:false" json:"-"`
	Difficulty   string `gorm:"type:difficulty_level;default:'easy'"`
}
type PhraseList struct {
	gorm.Model
	MaterialID     uint     `gorm:"type:uint;not null;index"`
	Title          string   `gorm:"type:varchar(255);not null"`
	Phrases        []Phrase `gorm:"foreignKey:PhraseListID;constraint:OnDelete:CASCADE"`
	GenerateStatus string   `gorm:"type:phrase_list_status;default:'pending'"`
}
