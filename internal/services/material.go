package services

import (
	"errors"
	"fmt"
	"sync"

	"newln/internal/models"
	"newln/internal/stores"

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
}

type materialService struct {
	store stores.MaterialStore
	mu    sync.Mutex
}

func NewMaterialService(s stores.MaterialStore) MaterialService {
	return &materialService{store: s}
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
	if ulid != material.LocalULID {
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
