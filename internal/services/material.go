package services

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yomek33/newln/internal/logger"
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
	UpdateHasPendingWordStatus(ulid string, status bool) error
	UpdateHasPendingPhraseStatus(ulid string, status bool) error
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
	logger.Infof("📡 Sending WebSocket update for %s: %s", materialULID, message)

	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers, ok := s.subscribers[materialULID]
	if !ok {
		logger.Warnf("⚠️ No WebSocket subscribers for materialULID: %s", materialULID)
		return
	}

	for _, ch := range subscribers {
		select {
		case ch <- message:
			logger.Infof("✅ Sent WebSocket update: %s", message)
		default:
			logger.Warnf("⚠️ WebSocket channel full, skipping: %s", materialULID)
		}
	}
}


func (s *materialService) UpdateMaterialField(ulid string, field string, value interface{}) error {
	logger.Infof("🛠 Updating field %s to %v for material %s", field, value, ulid)
	err := s.store.UpdateMaterialField(ulid, field, value)
	if err != nil {
		logger.Errorf("❌ Failed to update %s for material %s: %v", field, ulid, err)
	}
	return err
}


func (s *materialService) UpdateHasPendingWordStatus(ulid string, status bool) error {
	logger.Infof("🛠 Updating hasPendingWordStatus to %v for material %s", status, ulid)
	err := s.store.UpdateHasPendingWordStatus(ulid, status)
	if err != nil {
		logger.Errorf("❌ Failed to update hasPendingWordStatus for material %s: %v", ulid, err)
	}
	return err
}

func (s *materialService) UpdateHasPendingPhraseStatus(ulid string, status bool) error {
	logger.Infof("🛠 Updating hasPendingPhraseStatus to %v for material %s", status, ulid)
	err := s.store.UpdateHasPendingPhraseStatus(ulid, status)
	if err != nil {
		logger.Errorf("❌ Failed to update hasPendingPhraseStatus for material %s: %v", ulid, err)
	}
	return err
}