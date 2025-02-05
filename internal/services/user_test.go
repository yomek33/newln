package services_test

import (
	"errors"
	"testing"

	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/services"

	"github.com/stretchr/testify/assert"
)

type mockUserStore struct {
	users map[string]*models.User
}

func (m *mockUserStore) CreateUser(user *models.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("user already exists")
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserStore) GetUserByEmail(email string) (*models.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func TestUserService(t *testing.T) {
	mockStore := &mockUserStore{users: make(map[string]*models.User)}
	userService := services.NewUserService(mockStore)

	t.Run("RegisterUser", func(t *testing.T) {
		err := userService.RegisterUser("test@example.com", "password123", "Test User")
		assert.NoError(t, err)

		// 重複するユーザーを登録
		err = userService.RegisterUser("test@example.com", "password123", "Test User")
		assert.Error(t, err)
	})

	t.Run("LoginUser", func(t *testing.T) {
		_, err := userService.LoginUser("test@example.com", "wrongpassword")
		assert.Error(t, err)

		_, err = userService.LoginUser("test@example.com", "password123")
		assert.NoError(t, err)
	})
}
