package models

import (
	"gorm.io/gorm"
)

type Phrase struct {
	gorm.Model
	MaterialID uint   `gorm:"not null;index"`
	Text       string `gorm:"type:varchar(255);not null"`
	Importance string `gorm:"type:importance_level;default:'medium'"`
}
