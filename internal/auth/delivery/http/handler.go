package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/AriartyyyA/gobank/internal/auth/delivery/http/dto"
	"github.com/AriartyyyA/gobank/internal/auth/domain"
	"github.com/AriartyyyA/gobank/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type AuthUserCase interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
}

type HandlerAuth struct {
	uc        AuthUserCase
	validate  *validator.Validate
	jwtSecret string
}

func NewHandlerAuth(uc AuthUserCase, jwtSecret string) *HandlerAuth {
	return &HandlerAuth{
		uc:        uc,
		validate:  validator.New(),
		jwtSecret: jwtSecret,
	}
}

func (h *HandlerAuth) RegisterRoutes(router chi.Router) {
	// Публичные роуты, без мидлвари
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)

	// защищенные роуты
	router.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware(h.jwtSecret))
		r.Get("/users/me", h.Me)
	})
}

func (h *HandlerAuth) Register(w http.ResponseWriter, r *http.Request) {
	var reqDto dto.LoginAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "Incorrect data")
		return
	}

	if err := h.validate.Struct(reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "Bad email or password")
		return
	}

	if err := h.uc.Register(r.Context(), reqDto.Email, reqDto.Password); err != nil {
		log.Printf("register error: %v", err)
		if errors.Is(err, domain.ErrUserExists) {
			respondError(w, http.StatusConflict, "User exists")
			return
		}

		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "User created"})
}

func (h *HandlerAuth) Login(w http.ResponseWriter, r *http.Request) {
	var reqDto dto.LoginAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "Incorrect data")
		return
	}

	if err := h.validate.Struct(reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "Bad email or password")
		return
	}

	token, err := h.uc.Login(r.Context(), reqDto.Email, reqDto.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		if errors.Is(err, domain.ErrWrongPassword) {
			respondError(w, http.StatusUnauthorized, "Incorrect password or email")
			return
		}

		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	resp := dto.LoginResponse{
		Token: token,
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *HandlerAuth) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	respondJSON(w, http.StatusOK, map[string]string{"user_id": userID})
}
