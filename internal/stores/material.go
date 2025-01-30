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
	CreateMaterial(material *models.Material) (*models.Material, error)
	GetMaterialByID(ulid string, UserID uuid.UUID) (*models.Material, error)
	UpdateMaterial(ulid string, material *models.Material) error
	DeleteMaterial(ulid string, UserID uuid.UUID) error
	GetAllMaterials(searchQuery string, UserID uuid.UUID) ([]models.Material, error)
	UpdateMaterialStatus(ulid string, status string) error
	GetMaterialStatus(ulid string) (string, error)
}

type materialStore struct {
	DB *gorm.DB
}

func NewMaterialStore(db *gorm.DB) MaterialStore {
	return &materialStore{DB: db}
}

func (s *materialStore) CreateMaterial(material *models.Material) (*models.Material, error) {
	if material == nil {
		return nil, errors.New(ErrMaterialCannotBeNil)
	}
	err := s.DB.Create(material).Error
	if err != nil {
		return nil, err
	}
	return material, nil
}

func (s *materialStore) GetMaterialByID(ulid string, UserID uuid.UUID) (*models.Material, error) {
	log.Println("store material id", ulid)
	var material models.Material
	err := s.DB.Where("local_ul_id = ? AND user_id = ?", ulid, UserID).Preload("Phrases").Preload("Chats").First(&material).Error
	return &material, err
}

func (s *materialStore) UpdateMaterial(ulid string, material *models.Material) error {
	if material == nil {
		return errors.New(ErrMaterialCannotBeNil)
	}
	if ulid != material.LocalULID {
		return errors.New(ErrMismatchedMaterialID)
	}
	return s.DB.Model(&models.Material{}).Where("id = ?", ulid).Updates(material).Error
}

func (s *materialStore) DeleteMaterial(ulid string, UserID uuid.UUID) error {
	return s.DB.Where("local_ul_id = ? AND user_id = ?", ulid, UserID).Delete(&models.Material{}).Error
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

func (s *materialStore) UpdateMaterialStatus(ulid string, status string) error {
	if status != models.StatusDraft && status != models.StatusArchived && status != models.StatusPublished {
		return errors.New("invalid status")
	}
	return s.DB.Model(&models.Material{}).Where("local_ul_id = ?", ulid).Update("status", status).Error
}

func (s *materialStore) GetMaterialStatus(ulid string) (string, error) {
	var material models.Material
	err := s.DB.Select("status").Where("local_ul_id = ?", ulid).First(&material).Error
	return material.Status, err
}
