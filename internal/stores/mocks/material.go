package stores_mock

import (
	"errors"

	"github.com/google/uuid"
	"github.com/yomek33/newln/internal/models"
)

type MockMaterialStore struct {
	Materials map[uint]*models.Material
}

func NewMockMaterialStore() *MockMaterialStore {
	return &MockMaterialStore{
		Materials: make(map[uint]*models.Material),
	}
}

func (m *MockMaterialStore) CreateMaterial(material *models.Material) (*models.Material, error) {
	m.Materials[material.ID] = material
	return material, nil
}

func (m *MockMaterialStore) GetMaterialByULID(ulid string, UserID uuid.UUID) (*models.Material, error) {
	for _, material := range m.Materials {
		if material.ULID == ulid && material.UserID == UserID {
			return material, nil
		}
	}
	return nil, errors.New("material not found")
}

func (m *MockMaterialStore) GetMaterialByID(id uint) (*models.Material, error) {
	if material, exists := m.Materials[id]; exists {
		return material, nil
	}
	return nil, errors.New("material not found")
}

func (m *MockMaterialStore) UpdateMaterial(ulid string, material *models.Material) error {
	for id, mat := range m.Materials {
		if mat.ULID == ulid {
			m.Materials[id] = material
			return nil
		}
	}
	return errors.New("material not found")
}

func (m *MockMaterialStore) DeleteMaterial(ulid string, UserID uuid.UUID) error {
	for id, material := range m.Materials {
		if material.ULID == ulid && material.UserID == UserID {
			delete(m.Materials, id)
			return nil
		}
	}
	return errors.New("material not found")
}

func (m *MockMaterialStore) GetAllMaterials(searchQuery string, UserID uuid.UUID) ([]models.Material, error) {
	var materials []models.Material
	for _, material := range m.Materials {
		if material.UserID == UserID {
			materials = append(materials, *material)
		}
	}
	return materials, nil
}

func (m *MockMaterialStore) UpdateMaterialStatus(id uint, status string) error {
	if material, exists := m.Materials[id]; exists {
		material.Status = status
		return nil
	}
	return errors.New("material not found")
}

func (m *MockMaterialStore) GetMaterialStatus(ulid string) (string, error) {
	for _, material := range m.Materials {
		if material.ULID == ulid {
			return material.Status, nil
		}
	}
	return "", errors.New("material not found")
}

func (m *MockMaterialStore) CheckAllCompleted(materialID uint) (bool, error) {
	if _, exists := m.Materials[materialID]; exists {
		return true, nil
	}
	return false, errors.New("material not found")
}
