package services

import "newln/internal/stores"

type Services struct {
	UserService UserService
}

func NewServices(stores *stores.Stores) *Services {
	return &Services{
		UserService: NewUserService(stores.UserStore),
	}
}
