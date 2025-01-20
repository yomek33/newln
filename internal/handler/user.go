package handler

import (
	"net/http"
	"newln/internal/services"

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

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Validation failed"})
	}

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
