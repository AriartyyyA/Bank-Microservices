package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/AriartyyyA/gobank/internal/wallet/domain"
	"github.com/google/uuid"
)

type WalletUseCase struct {
	repo domain.WalletRepository
}

func NewWalletUseCase(repo domain.WalletRepository) *WalletUseCase {
	return &WalletUseCase{
		repo: repo,
	}
}

func (uc *WalletUseCase) CreateWallet(ctx context.Context, userID string) error {
	id := uuid.New().String()

	wallet := domain.Wallet{
		ID:        id,
		UserID:    userID,
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.repo.CreateWallet(ctx, wallet); err != nil {
		return fmt.Errorf("create wallet: %w", err)
	}

	return nil
}
