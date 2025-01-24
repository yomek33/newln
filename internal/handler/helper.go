package handler

import (
	"errors"
	"net/http"
	"newln/internal/logger"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("your-secret-key")
func getUserIDFromContext(c echo.Context) (string, error) {
    user, ok := c.Get("UserID").(string)
    if !ok || user == "" {
        logger.Errorf(ErrInvalidUserToken)
        return "", errors.New(ErrInvalidUserToken)
    }
    return user, nil // 有効なUserIDを返す
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