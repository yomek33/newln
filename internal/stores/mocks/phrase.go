package stores_mock

import (
	"errors"

	"github.com/yomek33/newln/internal/models"
)

type MockPhraseStore struct {
    Phrases     map[uint]*models.Phrase
    PhraseLists map[uint]*models.PhraseList
}

func NewMockPhraseStore() *MockPhraseStore {
    return &MockPhraseStore{
        Phrases:     make(map[uint]*models.Phrase),
        PhraseLists: make(map[uint]*models.PhraseList),
    }
}

func (m *MockPhraseStore) CreatePhrase(phrase *models.Phrase, phraseList *models.PhraseList) error {
    if phrase == nil || phraseList == nil {
        return errors.New("phrase or phraseList cannot be nil")
    }
    m.Phrases[phrase.ID] = phrase
    m.PhraseLists[phraseList.ID] = phraseList
    return nil
}

func (m *MockPhraseStore) GetPhrasesByMaterialID(materialULID string) ([]models.Phrase, error) {
    var phrases []models.Phrase
    for _, phrase := range m.Phrases {
            phrases = append(phrases, *phrase)
        
    }
    return phrases, nil
}

func (m *MockPhraseStore) CreatePhraseList(phraseList *models.PhraseList) error {
    if phraseList == nil {
        return errors.New("phraseList cannot be nil")
    }
    m.PhraseLists[phraseList.ID] = phraseList
    return nil
}

func (m *MockPhraseStore) GetPhraseListByMaterialULID(materialULID string) ([]models.PhraseList, error) {
    var phraseLists []models.PhraseList
    for _, phraseList := range m.PhraseLists {
            phraseLists = append(phraseLists, *phraseList)
        
    }
    return phraseLists, nil
}

func (m *MockPhraseStore) BulkInsertPhrases(phrases []models.Phrase) error {
    for _, phrase := range phrases {
        m.Phrases[phrase.ID] = &phrase
    }
    return nil
}

func (m *MockPhraseStore) UpdatePhraseListGenerateStatus(phraseListID uint, status string) error {
    if phraseList, exists := m.PhraseLists[phraseListID]; exists {
        phraseList.GenerateStatus = status
        return nil
    }
    return errors.New("phraseList not found")
}