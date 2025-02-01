package stores

import (
	"gorm.io/gorm"
)

type Stores struct {
	DB            *gorm.DB
	UserStore     UserStore
	MaterialStore MaterialStore
	PhraseStore   PhraseStore
	WordStore    WordStore
}

func NewStores(db *gorm.DB) *Stores {
	return &Stores{
		DB:            db,
		UserStore:     NewUserStore(db),
		MaterialStore: NewMaterialStore(db),
		PhraseStore:   NewPhraseStore(db),
		WordStore:    NewWordStore(db),
	}
}
