package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"newln/internal/services"
	"strings"
	"unicode"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	Service services.UserService
}

func NewUserHandler(s services.UserService) *UserHandler {
	return &UserHandler{Service: s}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"required"`
}

func (h *UserHandler) RegisterUser(c echo.Context) error {
	type RegisterRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
		Name     string `json:"name" validate:"required"`
	}

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request"})
	}

	// Email のバリデーション
	if err := validateEmail(req.Email); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	// Password のバリデーション
	if err := validatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	// サービス層での処理
	err := h.Service.RegisterUser(req.Email, req.Password, req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User registered successfully"})
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *UserHandler) LoginUser(c echo.Context) error {

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request"})
	}

	token, err := h.Service.LoginUser(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

//　util
func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	//TODO: prodではもっと強固なパスワードのバリデーションを行う
	var (
		hasMinLen   = false
		hasUpper    = false
		hasLower    = false
		hasNumber   = false
		hasSpecial  = false
		minLength   = 8
		specialChar = `!@#$%^&*()-_=+[]{}|;:'",.<>?/`
	)

	if len(password) >= minLength {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case strings.ContainsRune(specialChar, char):
			hasSpecial = true
		}
	}

	if !hasMinLen {
		return errors.New("password must be at least 8 characters long")
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}