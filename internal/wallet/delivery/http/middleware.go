package http

import (
	"context"
	"net/http"
	"strings"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (userID, email string, err error)
}

type contextKey string

const UserIDKey contextKey = "userID"

func GRPCAuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				respondError(w, http.StatusUnauthorized, "missing token")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			userID, _, err := validator.ValidateToken(r.Context(), tokenStr)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "validate error")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
