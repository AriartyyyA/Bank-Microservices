package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AriartyyyA/gobank/internal/wallet/delivery/http/dto"
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
	var transfer dto.TransferRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&transfer); err != nil {
		respondError(w, http.StatusBadRequest, "decoding error")
		return
	}

	err := h.uc.Transfer(r.Context(), transfer.FromWalletID, transfer.ToWalletID, transfer.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNegativeAmount) {
			respondError(w, http.StatusBadRequest, "negative amount")
			return
		}
		if errors.Is(err, domain.ErrToSendMyself) {
			respondError(w, http.StatusBadRequest, "try to send myself")
			return
		}
		if errors.Is(err, domain.ErrNoMoney) {
			respondError(w, http.StatusBadRequest, "no money")
			return
		}
		if errors.Is(err, domain.ErrFailedToUpdateBalance) {
			respondError(w, http.StatusInternalServerError, "update balance error")
			return
		}

		respondError(w, http.StatusInternalServerError, "server err")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"transfer status": "ok"})
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
