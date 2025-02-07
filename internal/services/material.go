package services

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/stores"

	"github.com/google/uuid"
)

type MaterialService interface {
	CreateMaterial(material *models.Material) (*models.Material, error)
	GetMaterialByULID(ulid string, UserID uuid.UUID) (*models.Material, error)
	UpdateMaterial(ulid string, material *models.Material) error
	DeleteMaterial(ulid string, UserID uuid.UUID) error
	GetAllMaterials(searchQuery string, UserID uuid.UUID) ([]models.Material, error)
	UpdateMaterialStatus(materialID uint, status string) error
	GetMaterialStatus(ulid string) (string, error)
	SubscribeToMaterialUpdates(materialULID string) chan string
	UnsubscribeFromMaterialUpdates(materialULID string, ch chan string)
	PublishMaterialUpdate(materialULID string, message string)
	UpdateMaterialField(ulid string, field string, value interface{}) error
}

type materialService struct {
	store       stores.MaterialStore
	mu          sync.Mutex
	subscribers map[string][]chan string //sse用
}

func NewMaterialService(s stores.MaterialStore) MaterialService {
	return &materialService{
		store:       s,
		subscribers: make(map[string][]chan string),
	}
}

var (
	ErrMaterialNil          = errors.New("material cannot be nil")
	ErrMismatchedMaterialID = errors.New("mismatched material ID")
)

func (s *materialService) CreateMaterial(material *models.Material) (*models.Material, error) {
	if material == nil {
		return nil, errors.New("material cannot be nil")
	}
	return s.store.CreateMaterial(material)
}

func (s *materialService) GetMaterialByULID(ulid string, UserID uuid.UUID) (*models.Material, error) {
	material, err := s.store.GetMaterialByULID(ulid, UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get material by ID: %w", err)
	}
	return material, nil
}

func (s *materialService) UpdateMaterial(ulid string, material *models.Material) error {
	if material == nil {
		return ErrMaterialNil
	}
	if ulid != material.ULID {
		return ErrMismatchedMaterialID
	}
	return s.store.UpdateMaterial(ulid, material)
}

func (s *materialService) DeleteMaterial(ulid string, UserID uuid.UUID) error {
	return s.store.DeleteMaterial(ulid, UserID)
}

func (s *materialService) GetAllMaterials(searchQuery string, UserID uuid.UUID) ([]models.Material, error) {
	return s.store.GetAllMaterials(searchQuery, UserID)
}

func (s *materialService) UpdateMaterialStatus(id uint, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.UpdateMaterialStatus(id, status)
}

func (s *materialService) GetMaterialStatus(ulid string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.GetMaterialStatus(ulid)
}

// 🔥 SSE用の購読機能
func (s *materialService) SubscribeToMaterialUpdates(materialULID string) chan string {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan string, 10)
	s.subscribers[materialULID] = append(s.subscribers[materialULID], ch)

	return ch
}

func (s *materialService) UnsubscribeFromMaterialUpdates(materialULID string, ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subscribers == nil {
		s.subscribers = make(map[string][]chan string)
	}

	channels, exists := s.subscribers[materialULID]
	if !exists {
		return // 登録されていない場合は何もしない
	}

	newChannels := make([]chan string, 0, len(channels))
	for _, c := range channels {
		if c != ch {
			newChannels = append(newChannels, c)
		}
	}

	//  チャネルリストを更新
	if len(newChannels) == 0 {
		delete(s.subscribers, materialULID)
	} else {
		s.subscribers[materialULID] = newChannels
	}

	close(ch)
}
func (s *materialService) PublishMaterialUpdate(materialULID string, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers, ok := s.subscribers[materialULID]
	if !ok {
		return // 登録されていなければ何もしない
	}

	for _, ch := range subscribers {
		select {
		case ch <- message:
		default:
			//  送信できない場合はスキップ（クライアントが切断済み）
		}
	}
}


func (s *materialService) UpdateMaterialField(ulid string, field string, value interface{}) error {
	return s.store.UpdateMaterialField(ulid, field, value)
}