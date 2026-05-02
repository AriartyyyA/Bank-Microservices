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

func (uc *WalletUseCase) Transfer(ctx context.Context, fromWalletID, toWalletID string, amount int64) error {
	if amount < 0 {
		return domain.ErrTransactionNegativeAmount
	}

	if fromWalletID == toWalletID {
		return domain.ErrToSendMyself
	}

	walletFrom, err := uc.repo.FindWalletByID(ctx, fromWalletID)
	if err != nil {
		return fmt.Errorf("find wallet: %w", err)
	}

	if walletFrom.Balance < amount {
		return domain.ErrNoMoney
	}

	_, err = uc.repo.FindWalletByID(ctx, toWalletID)
	if err != nil {
		return fmt.Errorf("find wallet: %w", err)
	}

	return uc.repo.WithTx(ctx, func(ctx context.Context) error {
		if err := uc.repo.UpdateBalance(ctx, fromWalletID, -amount); err != nil {
			return fmt.Errorf("update balance from: %w", err)
		}

		if err := uc.repo.UpdateBalance(ctx, toWalletID, amount); err != nil {
			return fmt.Errorf("update balance to: %w", err)
		}

		id := uuid.New().String()

		transaction := domain.Transaction{
			ID:           id,
			FromWalletID: fromWalletID,
			ToWalletID:   toWalletID,
			Amount:       amount,
			CreatedAt:    time.Now(),
		}

		if err := uc.repo.CreateTransaction(ctx, transaction); err != nil {
			return err
		}

		return nil
	})
}

func (uc *WalletUseCase) GetBalance(ctx context.Context, walletID string) (int64, error) {
	wallet, err := uc.repo.FindWalletByID(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}

	return wallet.Balance, nil
}

func (uc *WalletUseCase) GetBalanceByUserID(ctx context.Context, userID string) (int64, error) {
	wallet, err := uc.repo.FindWalletByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}

	return wallet.Balance, nil
}

func (uc *WalletUseCase) GetHistory(ctx context.Context, walletID string) ([]*domain.Transaction, error) {
	transactions, err := uc.repo.GetTransactionsByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	return transactions, nil
}
