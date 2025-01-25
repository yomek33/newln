package stores

import (
	"errors"

	"newln/internal/models"

	"gorm.io/gorm"
)

type PhraseStore interface {
	CreatePhrase(phrase *models.Phrase) error
	GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error)
}

type phraseStore struct {
	DB *gorm.DB
}

func NewPhraseStore(db *gorm.DB) PhraseStore {
	return &phraseStore{DB: db}
}

func (s *phraseStore) CreatePhrase(phrase *models.Phrase) error {
	if phrase == nil {
		return errors.New("phrase cannot be nil")
	}

	if phrase.Text == "" {
		return errors.New("phrase Text cannot be empty")
	}
	if phrase.MaterialID == 0 {
		return errors.New("phrase MaterialID cannot be empty")
	}

	return s.DB.Create(phrase).Error
}

func (s *phraseStore) GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error) {
	var phrases []models.Phrase
	err := s.DB.Where("material_id = ?", materialULID).Find(&phrases).Error
	return phrases, err
}
