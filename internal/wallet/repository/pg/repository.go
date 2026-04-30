package pg_repo

import (
	"context"

	"github.com/AriartyyyA/gobank/internal/wallet/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ domain.WalletRepository = (*PostgresRepo)(nil)

type PostgresRepo struct {
	connPool *pgxpool.Pool
}

func NewPostgresRepo(connPool *pgxpool.Pool) domain.WalletRepository {
	return &PostgresRepo{
		connPool: connPool,
	}
}

func (r *PostgresRepo) CreateWallet(ctx context.Context, wallet domain.Wallet) error {
	return nil
}

func (r *PostgresRepo) FindWalletByID(ctx context.Context, walletID string) (*domain.Wallet, error) {
	return nil, nil
}

func (r *PostgresRepo) FindWalletByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	return nil, nil
}

func (r *PostgresRepo) UpdateBalance(ctx context.Context, walletID string, amount int64) error {
	return nil
}

func (r *PostgresRepo) GetTransactionsByWalletID(ctx context.Context, walletID string) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *PostgresRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return nil
}

func (r *PostgresRepo) CreateTransaction(ctx context.Context, transaction domain.Transaction) error {
	return nil
}
