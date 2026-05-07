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
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	ValidateToken(token string) (userID, email string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
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
	router.Post("/auth/refresh", h.RefreshToken)
	router.Post("/auth/logout", h.Logout)

	// защищенные роуты
	router.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware(h.jwtSecret))
		r.Get("/users/me", h.Me)
	})
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginAndRegisterRequest true "Данные пользователя"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/register [post]
func (h *HandlerAuth) Register(w http.ResponseWriter, r *http.Request) {
	var reqDto dto.LoginAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "Incorrect data")
		return
	}

	if err := h.validate.Struct(reqDto); err != nil {
		if respondValidationError(w, err) {
			return
		}
		respondError(w, http.StatusBadRequest, "Invalid request")
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

// Login godoc
// @Summary      Вход пользователя
// @Description  Авторизирует пользователя и выдает jwt-токен
// @Tags         login
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginAndRegisterRequest true "Данные пользователя"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/login [post]
func (h *HandlerAuth) Login(w http.ResponseWriter, r *http.Request) {
	var reqDto dto.LoginAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "Incorrect data")
		return
	}

	if err := h.validate.Struct(reqDto); err != nil {
		if respondValidationError(w, err) {
			return
		}
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	accessToken, refreshToken, err := h.uc.Login(r.Context(), reqDto.Email, reqDto.Password)
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *HandlerAuth) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	respondJSON(w, http.StatusOK, map[string]string{"user_id": userID})
}

func (h *HandlerAuth) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Incorrect data")
		return
	}

	accessToken, err := h.uc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			respondError(w, http.StatusBadRequest, "invalid refresh token")
			return
		}

		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusOK, dto.LoginResponse{AccessToken: accessToken})
}

func (h *HandlerAuth) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Incorrect data")
		return
	}

	if err := h.uc.Logout(r.Context(), req.RefreshToken); err != nil {
		respondError(w, http.StatusBadRequest, "logout error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "logout"})
}

func messageForTag(tag string) string {
	switch tag {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "value is too short"
	case "max":
		return "value is too long"
	default:
		return "invalid value"
	}
}

func respondValidationError(w http.ResponseWriter, err error) bool {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make([]map[string]string, 0)
		for _, e := range validationErrors {
			details = append(details, map[string]string{
				"field":   e.Field(),
				"message": messageForTag(e.Tag()),
			})
		}
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":   "validation failed",
			"details": details,
		})
		return true
	}
	return false
}
