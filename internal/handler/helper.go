package handler

import (
	"errors"
	"net/http"
	"newln/internal/logger"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)



var secretKey = []byte("your-secret-key")

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, ErrUnauthorized, http.StatusUnauthorized)
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return secretKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, ErrUnauthorized, http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func getUserIDFromContext(c echo.Context) (string, error) {
    user, ok := c.Get("UserID").(string)
    if !ok || user == "" {
        logger.Errorf(ErrInvalidUserToken)
        return "", errors.New(ErrInvalidUserToken)
    }
    return user, nil // 有効なUserIDを返す
}

var jwtSecret = []byte("your-secret-key")

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        authHeader := c.Request().Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            return echo.NewHTTPError(http.StatusUnauthorized, ErrMissingOrInvalidToken)
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidToken)
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || claims["sub"] == nil {
            return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidClaims)
        }

        userID := claims["sub"].(string)
        c.Set("UserID", userID) // コンテキストにUserIDを格納

        return next(c)
    }
}