package stores

import (
	"errors"
	"log"

	"newln/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
    ErrMaterialCannotBeNil  = "material cannot be nil"
    ErrMismatchedMaterialID = "material ID does not match the provided ID"
)

type MaterialStore interface {
    CreateMaterial(material *models.Material) (uint, error)
    GetMaterialByID(id uint, UserID uuid.UUID) (*models.Material, error)
    UpdateMaterial(id uint, material *models.Material) error
    DeleteMaterial(id uint, UserID uuid.UUID) error
    GetAllMaterials(searchQuery string, UserID uuid.UUID) ([]models.Material, error)
    UpdateMaterialStatus(id uint, status string) error
    GetMaterialStatus(id uint) (string, error)
}

type materialStore struct {
    DB *gorm.DB
}

func NewMaterialStore(db *gorm.DB) MaterialStore {
	return &materialStore{DB: db}
}

func (s *materialStore) CreateMaterial(material *models.Material) (uint, error) {
    if material == nil {
        return 0, errors.New(ErrMaterialCannotBeNil)
    }
    err := s.DB.Create(material).Error
    if err != nil {
        return 0, err
    }
    return material.ID, nil
}

func (s *materialStore) GetMaterialByID(id uint, UserID uuid.UUID) (*models.Material, error) {
    log.Println("store material id", id)
    var material models.Material
    err := s.DB.Where("id = ? AND user_id = ?", id, UserID).Preload("Phrases").Preload("Chats").First(&material).Error
    return &material, err
}

func (s *materialStore) UpdateMaterial(id uint, material *models.Material) error {
    if material == nil {
        return errors.New(ErrMaterialCannotBeNil)
    }
    if id != material.ID {
        return errors.New(ErrMismatchedMaterialID)
    }
    return s.DB.Model(&models.Material{}).Where("id = ?", id).Updates(material).Error
}

func (s *materialStore) DeleteMaterial(id uint, UserID uuid.UUID) error {
    return s.DB.Where("id = ? AND user_id = ?", id, UserID).Delete(&models.Material{}).Error
}

func (s *materialStore) GetAllMaterials(searchQuery string, UserID uuid.UUID) ([]models.Material, error) {
    var materials []models.Material
    query := s.DB.Where("user_id = ?", UserID)
    if searchQuery != "" {
        query = query.Where("title LIKE ?", "%"+searchQuery+"%")
    }
    err := query.Find(&materials).Error
    return materials, err
}

func (s *materialStore) UpdateMaterialStatus(id uint, status string) error {
    if status != models.StatusDraft && status != models.StatusArchived && status != models.StatusPublished {
        return errors.New("invalid status")
    }
    return s.DB.Model(&models.Material{}).Where("id = ?", id).Update("status", status).Error
}

func (s *materialStore) GetMaterialStatus(id uint) (string, error) {
    var material models.Material
    err := s.DB.Select("status").Where("id = ?", id).First(&material).Error
    return material.Status, err
}