package stores

import (
	"errors"

	"newln/internal/models"

	"gorm.io/gorm"
)

type WordStore interface {
    CreateWord(word *models.Word, wordList *models.WordList) error
    GetWordsByMaterialID(materialULID string) ([]models.Word, error)
    CreateWordList(wordList *models.WordList) error
    GetWordListByMaterialULID(materialULID string) ([]models.WordList, error)
    BulkInsertWords(words []models.Word) error
	UpdateWordListGenerateStatus(wordListID uint, status string) error
}

type wordStore struct {
    DB *gorm.DB
}

func NewWordStore(db *gorm.DB) WordStore {
    return &wordStore{DB: db}
}

func (s *wordStore) CreateWordList(wordList *models.WordList) error {
    if wordList == nil {
        return errors.New("wordList cannot be nil")
    }
    if wordList.MaterialID == 0 {
        return errors.New("wordList MaterialID cannot be empty")
    }
    if wordList.Title == "" {
        return errors.New("wordList Title cannot be empty")
    }

    return s.DB.Create(wordList).Error
}

func (s *wordStore) GetWordListByMaterialULID(materialULID string) ([]models.WordList, error) {
    var wordLists []models.WordList

    err := s.DB.Joins("JOIN materials ON materials.id = word_lists.material_id").
        Where("materials.local_ulid = ?", materialULID).
        Find(&wordLists).Error

    return wordLists, err
}

func (s *wordStore) CreateWord(word *models.Word, wordList *models.WordList) error {
    if word == nil {
        return errors.New("word cannot be nil")
    }
    if word.Text == "" {
        return errors.New("word Text cannot be empty")
    }
    if wordList == nil {
        return errors.New("wordList cannot be nil")
    }

    word.WordListID = wordList.ID
    return s.DB.Create(word).Error
}

func (s *wordStore) GetWordsByMaterialID(materialULID string) ([]models.Word, error) {
    var words []models.Word

    err := s.DB.Joins("JOIN word_lists ON word_lists.id = words.word_list_id").
        Joins("JOIN materials ON materials.id = word_lists.material_id").
        Where("materials.local_ulid = ?", materialULID).
        Find(&words).Error

    return words, err
}

func (s *wordStore) BulkInsertWords(words []models.Word) error {
    return s.DB.CreateInBatches(words, 100).Error
}

func (s *wordStore) UpdateWordListGenerateStatus(wordListID uint, status string) error {
	return s.DB.Model(&models.WordList{}).Where("id = ?", wordListID).Update("generate_status", status).Error
}