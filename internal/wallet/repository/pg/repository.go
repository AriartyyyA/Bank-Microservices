package pg_repo

import (
	"context"
	"errors"
	"fmt"

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
	var wallet domain.Wallet

	query := `SELECT id, user_id, balance, created_at, updated_at FROM wallets WHERE user_id = $1`

	row := r.connPool.QueryRow(ctx, query, userID)

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

func (r *PostgresRepo) UpdateBalance(ctx context.Context, walletID string, amount int64) error {
	query := `UPDATE wallets SET balance = balance + $1 WHERE id = $2`

	if _, err := r.connPool.Exec(ctx, query, amount, walletID); err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}

func (r *PostgresRepo) GetTransactionsByWalletID(ctx context.Context, walletID string) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *PostgresRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return nil
}

func (r *PostgresRepo) CreateTransaction(ctx context.Context, transaction domain.Transaction) error {
	query := `INSERT INTO transactions(id, from_wallet_id, to_wallet_id, amount) VALUES($1, $2, $3, $4)`

	if _, err := r.connPool.Exec(
		ctx,
		query,
		transaction.ID,
		transaction.FromWalletID,
		transaction.ToWalletID,
		transaction.Amount,
	); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	return nil
}
