package stores

import (
	"errors"

	"newln/internal/models"

	"gorm.io/gorm"
)

type PhraseStore interface {
	CreatePhrase(phrase *models.Phrase, phraseList *models.PhraseList) error
	GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error)
	CreatePhraseList(phraseList *models.PhraseList) error
	GetPhraseListByMaterialULID(materialULID string) ([]models.PhraseList, error)
	//GetOrCreatePhraseList(materialID uint, phraseList *models.PhraseList) (*models.PhraseList, error)
	BulkInsertPhrases(phrases []models.Phrase) error
	UpdatePhraseListGenerateStatus(phraseListID uint, status string) error
}

type phraseStore struct {
	DB *gorm.DB
}

func NewPhraseStore(db *gorm.DB) PhraseStore {
	return &phraseStore{DB: db}
}

// ✅ `PhraseList` を作成（構造体を引数にする）
func (s *phraseStore) CreatePhraseList(phraseList *models.PhraseList) error {
	if phraseList == nil {
		return errors.New("phraseList cannot be nil")
	}
	if phraseList.MaterialID == 0 {
		return errors.New("phraseList MaterialID cannot be empty")
	}
	if phraseList.Title == "" {
		return errors.New("phraseList Title cannot be empty")
	}

	return s.DB.Create(phraseList).Error
}

func (s *phraseStore) GetPhraseListByMaterialULID(materialULID string) ([]models.PhraseList, error) {
	var phraseLists []models.PhraseList

	err := s.DB.Joins("JOIN materials ON materials.id = phrase_lists.material_id").
		Where("materials.local_ulid = ?", materialULID).
		Find(&phraseLists).Error

	return phraseLists, err
}

// TODO: `PhraseList` が存在しない場合、新規作成するのがあっても良い
// func (s *phraseStore) GetOrCreatePhraseList(materialID uint, title string) (*models.PhraseList, error) {
// 	var phraseList models.PhraseList

// 	// 既存の PhraseList を検索
// 	err := s.DB.Where("material_id = ? AND title = ?", materialID, title).First(&phraseList).Error
// 	if err == nil {
// 		return &phraseList, nil
// 	}

// 	// 存在しない場合は新規作成
// 	if errors.Is(err, gorm.ErrRecordNotFound) {
// 		newPhraseList := &models.PhraseList{
// 			MaterialID: materialID,
// 			Title:      title,
// 		}
// 		if err := s.CreatePhraseList(newPhraseList); err != nil {
// 			return nil, err
// 		}
// 		return newPhraseList, nil
// 	}

// 	return nil, err
// }

func (s *phraseStore) CreatePhrase(phrase *models.Phrase, phraseList *models.PhraseList) error {
	if phrase == nil {
		return errors.New("phrase cannot be nil")
	}
	if phrase.Text == "" {
		return errors.New("phrase Text cannot be empty")
	}
	if phraseList == nil {
		return errors.New("phraseList cannot be nil")
	}

	phrase.PhraseListID = phraseList.ID
	return s.DB.Create(phrase).Error
}

func (s *phraseStore) GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error) {
	var phrases []models.Phrase

	err := s.DB.Joins("JOIN phrase_lists ON phrase_lists.id = phrases.phrase_list_id").
		Joins("JOIN materials ON materials.id = phrase_lists.material_id").
		Where("materials.local_ulid = ?", materialULID).
		Find(&phrases).Error

	return phrases, err
}

func (s *phraseStore) BulkInsertPhrases(phrases []models.Phrase) error {
	return s.DB.CreateInBatches(phrases, 100).Error
}

func (s *phraseStore) UpdatePhraseListGenerateStatus(phraseListID uint, status string) error {
	return s.DB.Model(&models.PhraseList{}).Where("id = ?", phraseListID).Update("generate_status", status).Error
}
