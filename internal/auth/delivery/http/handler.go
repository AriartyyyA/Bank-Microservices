package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AriartyyyA/gobank/internal/auth/delivery/http/dto"
	"github.com/AriartyyyA/gobank/internal/auth/domain"
	"github.com/go-chi/chi/v5"
)

type AuthUserCase interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
}

type HandlerAuth struct {
	uc AuthUserCase
}

func NewHandlerAuth(uc AuthUserCase) *HandlerAuth {
	return &HandlerAuth{
		uc: uc,
	}
}

func (h *HandlerAuth) RegisterRoutes(router chi.Router) {
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
}

func (h *HandlerAuth) Register(w http.ResponseWriter, r *http.Request) {
	var reqDto dto.LoginAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.uc.Register(r.Context(), reqDto.Email, reqDto.Password); err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode("Пользователь успешно создан"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HandlerAuth) Login(w http.ResponseWriter, r *http.Request) {
	var reqDto dto.LoginAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.uc.Login(r.Context(), reqDto.Email, reqDto.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
			return
		}
		if errors.Is(err, domain.ErrWrongPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := dto.LoginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
