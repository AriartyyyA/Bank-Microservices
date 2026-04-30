package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

func (h *HandlerAuth) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Достаем сам токен
		authHeader := r.Header.Get("Authorization")
		// Проверяем если ли вообще токен и есть ли префикс
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		// Оставляем чистый токен, обрезая префикс
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return []byte(h.jwtSecret), nil
		})
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		userID := claims["userID"].(string)
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
