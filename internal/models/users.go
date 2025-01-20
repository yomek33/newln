package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
	UserID    string       `gorm:"primaryKey;type:varchar(255);not null;unique"`
	Name      string       `gorm:"type:varchar(255);not null"`
	Materials []Material   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Email     string       `gorm:"type:varchar(255);not null;unique"`
	Password  string       `gorm:"type:varchar(255);not null"`
}

type Progress struct {
	gorm.Model
	UserID       string    `gorm:"type:varchar(255);not null;index"`
	PhraseID     uint      `gorm:"not null;index"`
	Status       string    `gorm:"type:progress_status;not null"`
	LastReviewed time.Time `gorm:"not null"`
}
