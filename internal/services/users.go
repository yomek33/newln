package services

import (
	"errors"

	"github.com/yomek33/newln/internal/models"
	"github.com/yomek33/newln/internal/stores"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	RegisterUser(email, password, name string) error
	LoginUser(email, password string) (string, error)
}

type userService struct {
	Store stores.UserStore
}

func NewUserService(s stores.UserStore) UserService {
	return &userService{Store: s}
}

func (s *userService) RegisterUser(email, password, name string) error {
	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// 一意の UID を生成
	uid := uuid.New()

	user := &models.User{
		UserID:   uid,
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}

	return s.Store.CreateUser(user)
}

func (s *userService) LoginUser(email, password string) (string, error) {
	user, err := s.Store.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// パスワードの検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	// JWT トークンを生成するロジックを追加
	token := "generate_your_jwt_here" // TODO: トークン生成部分を実装する

	return token, nil
}
