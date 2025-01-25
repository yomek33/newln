package services

import "newln/internal/stores"

type Services struct {
	UserService UserService
	MaterialService MaterialService
	PhraseService PhraseService
}

func NewServices(stores *stores.Stores) *Services {
	return &Services{
		UserService: NewUserService(stores.UserStore),
		MaterialService: NewMaterialService(stores.MaterialStore),
		PhraseService: NewPhraseService(stores.PhraseStore, stores.MaterialStore),
	}
}
