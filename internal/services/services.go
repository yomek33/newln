package services

import (
	"github.com/yomek33/newln/internal/pkg/vertex"
	"github.com/yomek33/newln/internal/stores"
)

type Services struct {
	UserService     UserService
	MaterialService MaterialService
	PhraseService   PhraseService
	WordService     WordService
}

func NewServices(stores *stores.Stores, vertexService vertex.VertexService) *Services {
	return &Services{
		UserService:     NewUserService(stores.UserStore),
		MaterialService: NewMaterialService(stores.MaterialStore),
		PhraseService:   NewPhraseService(stores.PhraseStore, stores.MaterialStore, vertexService),
		WordService:     NewWordService(stores.WordStore, stores.MaterialStore, vertexService),
	}
}
