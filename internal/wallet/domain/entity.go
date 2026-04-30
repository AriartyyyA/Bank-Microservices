package domain

import (
	"context"
	"time"
)

type Wallet struct {
	ID        string
	UserID    string
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID           string
	FromWalletID string
	ToWalletID   string
	Amount       int64
	CreatedAt    time.Time
}

type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet Wallet) error
	FindWalletByID(ctx context.Context, walletID string) (*Wallet, error)
	FindWalletByUserID(ctx context.Context, userID string) (*Wallet, error)
	UpdateBalance(ctx context.Context, walletID string, amount int64) error
	CreateTransaction(ctx context.Context, transaction Transaction) error
	GetTransactionsByWalletID(ctx context.Context, walletID string) ([]*Transaction, error)
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
