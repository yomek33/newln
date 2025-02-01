package models

import "gorm.io/gorm"

type Word struct {
	gorm.Model
	WordListID uint   `gorm:"not null;index"`
	Text       string `gorm:"type:varchar(255);not null"`
	Importance string `gorm:"type:importance_level;default:'medium'"`
	Level      string `gorm:"type:word_level;default:'beginner'"`
	Meaning   string `gorm:"type:varchar(255)"`
}

type WordList struct {
    gorm.Model
    MaterialID uint `gorm:"type:uint;not null;index"`
    Title      string    `gorm:"type:varchar(255);not null"`
    Words      []Word    `gorm:"foreignKey:WordListID;constraint:OnDelete:CASCADE"`
	GenerateStatus string `gorm:"type:word_list_status;default:'pending'"`
}

