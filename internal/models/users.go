package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
	UserID    uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string       `gorm:"type:varchar(255);"`
	Materials []Material   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Email     string       `gorm:"type:varchar(255);unique"`
	Password  string       `gorm:"type:varchar(255)"`
}

type Progress struct {
	gorm.Model
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	PhraseID     uint      `gorm:"not null;index"`
	Status       string    `gorm:"type:progress_status;not null"`
	LastReviewed time.Time `gorm:"not null"`
}
