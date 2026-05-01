package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/AriartyyyA/gobank/internal/wallet/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type WalletUseCase interface {
	CreateWallet(ctx context.Context, userID string) error
	Transfer(ctx context.Context, fromWalletID, toWalletID string, amount int64) error
	GetBalance(ctx context.Context, walletID string) (int64, error)
	GetHistory(ctx context.Context, walletID string) ([]*domain.Transaction, error)
	GetBalanceByUserID(ctx context.Context, userID string) (int64, error)
}

type HandlerWallet struct {
	uc       WalletUseCase
	validate *validator.Validate
}

func NewHandlerWallet(uc WalletUseCase) *HandlerWallet {
	return &HandlerWallet{
		uc:       uc,
		validate: validator.New(),
	}
}

func (h *HandlerWallet) RegisterRoutes(router chi.Router) {
	router.Post("/wallets", h.CreateWallet)
	router.Get("/wallets/me", h.GetBalance)
	router.Post("/wallets/transfer", h.Transfer)
	router.Get("/wallet/history", h.GetHistory)
}

type contextKey string

const userIDKey contextKey = "userID"

func (h *HandlerWallet) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)

	if err := h.uc.CreateWallet(r.Context(), userID); err != nil {
		if errors.Is(err, domain.ErrWalletExists) {
			respondError(w, http.StatusConflict, "wallet already exists")
			return
		}

		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusCreated, "user created")
}

func (h *HandlerWallet) Transfer(w http.ResponseWriter, r *http.Request) {
}

func (h *HandlerWallet) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)

	balance, err := h.uc.GetBalanceByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			respondError(w, http.StatusNotFound, "wallet not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int64{"balance": balance})
}

func (h *HandlerWallet) GetHistory(w http.ResponseWriter, r *http.Request) {
}
