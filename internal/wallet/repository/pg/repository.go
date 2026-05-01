package pg_repo

import (
	"context"
	"errors"

	"github.com/AriartyyyA/gobank/internal/wallet/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Так можно проверять соответствие интерфейсу
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
	query := `INSERT INTO wallets(id, user_id, balance) VALUES ($1, $2, $3, $4, $5)`

	if _, err := r.connPool.Exec(
		ctx,
		query,
		wallet.ID,
		wallet.UserID,
		wallet.Balance,
	); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrWalletExists
		}

		return err
	}

	return nil
}

func (r *PostgresRepo) FindWalletByID(ctx context.Context, walletID string) (*domain.Wallet, error) {
	query := `SELECT id, user_id, balance, created_at, updated_at FROM wallets WHERE id = $1`

	row := r.connPool.QueryRow(ctx, query, walletID)

	var wallet domain.Wallet
	if err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}

		return nil, err
	}

	return &wallet, nil
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
