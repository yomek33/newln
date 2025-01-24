package handler

import (
	"errors"
	"fmt"
	"net/http"
	"newln/internal/logger"
	"newln/internal/models"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("your-secret-key")
func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
    userIDStr, ok := c.Get("UserID").(string)
    if !ok || userIDStr == "" {
        logger.Errorf(ErrInvalidUserToken)
        return uuid.Nil, errors.New(ErrInvalidUserToken)
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        logger.Errorf("Invalid UUID format: %v", err)
        return uuid.Nil, errors.New("invalid UUID format")
    }

    return userID, nil // 有効な UserID を返す
}

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return echo.NewHTTPError(http.StatusUnauthorized, ErrMissingOrInvalidToken)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidToken)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidClaims)
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidClaims)
		}

		c.Set("UserID", userID) // コンテキストにUserIDを格納

		return next(c)
	}
}

func respondWithError(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]string{"error": message})
}

func bindAndValidateMaterial(c echo.Context, material *models.Material) error {
	if err := c.Bind(material); err != nil {
		logger.Errorf("Error binding material: %v", err)
		return errors.New(ErrInvalidMaterialData)
	}
	if err := validateMaterial(material); err != nil {
		return err
	}
	return nil
}

func validateMaterial(material *models.Material) error {
	validate := validator.New()
	if err := validate.Struct(material); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessage := fmt.Sprintf("Error in field '%s': %s", strings.ToLower(err.Field()), err.Tag())
			errorMessages = append(errorMessages, errorMessage)

		}
		logger.Errorf("Error validating material: %v", errors.New(strings.Join(errorMessages, ", ")))
		return errors.New(strings.Join(errorMessages, ", "))
	}
	return nil
}

func parseUintParam(c echo.Context, paramName string) (uint, error) {
	param := c.Param(paramName)
	value, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		logger.Errorf("Error parsing uint param: %v", err)
		return 0, err
	}
	return uint(value), err
}
