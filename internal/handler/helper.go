package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/yomek33/newln/internal/logger"
	"github.com/yomek33/newln/internal/models"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userIDStr, ok := c.Get("UserID").(string)
	if !ok || userIDStr == "" {
		logger.Errorf(ErrInvalidUserToken)
		return uuid.Nil, errors.New(ErrInvalidUserToken)
	}
	logger.Infof("UserID: %s", userIDStr)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Errorf("Invalid UUID format: %v", err)
		return uuid.Nil, errors.New("invalid UUID format")
	}

	return userID, nil // 有効な UserID を返す
}
func (h *Handlers) JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("JWTMiddleware: missing or invalid token format")
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token format")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				fmt.Println("JWTMiddleware: unexpected signing method")
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return h.jwtSecret, nil
		})

		if err != nil {
			fmt.Printf("JWTMiddleware: JWT parse error: %v %v\n", err, token)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		if !token.Valid {
			fmt.Println("JWTMiddleware: invalid token")
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("JWTMiddleware: invalid claims structure")
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims structure")
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			fmt.Println("JWTMiddleware: missing or invalid 'sub' claim")
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid 'sub' claim")
		}

		// コンテキストにUserIDを格納
		c.Set("UserID", userID)

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
func isValidJWTToken(tokenString string, secretKey []byte) (string, error) {
	if tokenString == "" {
		return "", errors.New("missing token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		logger.Errorf("JWT parse error: %v", err)
		return "", errors.New("invalid token")
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims structure")
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", errors.New("missing or invalid 'sub' claim")
	}

	return userID, nil
}
