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
	CreateWallet(ctx context.Context, userID string) (*domain.Wallet, error)
	Transfer(ctx context.Context, fromWalletID, toWalletID string, amount int64) error
	GetBalance(ctx context.Context, walletID string) (int64, error)
	GetHistoryByUserID(ctx context.Context, userID string, limit, offset string) ([]*domain.Transaction, error)
	GetBalanceByUserID(ctx context.Context, userID string) (int64, error)
	UpdateBalance(ctx context.Context, userID string, amount int64) (int64, error)
	GetWalletByUserID(ctx context.Context, userID string) (*domain.Wallet, error)
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
	router.Post("/wallets/deposit", h.DepositBalance)
	router.Get("/wallet", h.GetWalletID)
}

// GetWallet godoc
// @Summary      Получение кошелька
// @Description  Возвращает информацию о кошельке по userID по JWT
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallet [get]
// @Security BearerAuth
func (h *HandlerWallet) GetWalletID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	wallet, err := h.uc.GetWalletByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			respondError(w, http.StatusNotFound, "wallet not found")
			return
		}

		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusOK, wallet)
}

// CreateWallet godoc
// @Summary      Создание кошелька
// @Description  Создаёт кошелек для пользователя
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallets [post]
// @Security BearerAuth
func (h *HandlerWallet) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	wallet, err := h.uc.CreateWallet(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletExists) {
			respondError(w, http.StatusConflict, "wallet already exists")
			return
		}

		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusCreated, wallet)
}

// DepositBalance godoc
// @Summary      Пополнение баланса
// @Description  Пополняем баланс пользователя
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Param        request body dto.DepositRequestDTO true "Данные пополнения"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallets/deposit [post]
// @Security BearerAuth
func (h *HandlerWallet) DepositBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	var reqDto dto.DepositRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&reqDto); err != nil {
		respondError(w, http.StatusBadRequest, "decoding error")
		return
	}

	amount, err := h.uc.UpdateBalance(r.Context(), userID, reqDto.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "update error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int64{"amount": amount})
}

// Transfer godoc
// @Summary      Создание перевода
// @Description  Создаем перевод с данными
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Param        request body dto.TransferRequestDTO true "Данные перевода"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallets/transfer [post]
// @Security BearerAuth
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

// GetBalance godoc
// @Summary      Получение баланса
// @Description  получаем баланс пользователя по id
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallets/me [get]
// @Security BearerAuth
func (h *HandlerWallet) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

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

// GetHistory godoc
// @Summary      История переводов
// @Description  Получаем историю переводов
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallet/history [get]
// @Security BearerAuth
func (h *HandlerWallet) GetHistory(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	limit := queryParams.Get("limit")
	offset := queryParams.Get("offset")

	userID := r.Context().Value(UserIDKey).(string)

	history, err := h.uc.GetHistoryByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			respondError(w, http.StatusNotFound, "wallet not found")
			return
		}

		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusOK, history)
}
