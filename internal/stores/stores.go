package stores

import (
	"gorm.io/gorm"
)

type Stores struct {
	DB        *gorm.DB
	UserStore UserStore
}

func NewStores(db *gorm.DB) *Stores {
	return &Stores{
		DB:        db,
		UserStore: NewUserStore(db),
	}
}
