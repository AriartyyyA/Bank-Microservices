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

type txKey struct {
}

func (r *PostgresRepo) getConn(ctx context.Context) interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
} {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return r.connPool
}

func NewPostgresRepo(connPool *pgxpool.Pool) domain.WalletRepository {
	return &PostgresRepo{
		connPool: connPool,
	}
}

func (r *PostgresRepo) CreateWallet(ctx context.Context, wallet domain.Wallet) error {
	query := `INSERT INTO wallets(id, user_id, balance) VALUES ($1, $2, $3)`

	if _, err := r.getConn(ctx).Exec(
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

	row := r.getConn(ctx).QueryRow(ctx, query, walletID)

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

	row := r.getConn(ctx).QueryRow(ctx, query, userID)

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

	if _, err := r.getConn(ctx).Exec(ctx, query, amount, walletID); err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}

func (r *PostgresRepo) GetTransactionsByWalletID(ctx context.Context, walletID string, limit, offset int) ([]*domain.Transaction, error) {
	query := `SELECT id, from_wallet_id, to_wallet_id, amount, created_at 
		FROM transactions
		WHERE from_wallet_id = $1 OR to_wallet_id = $1
		LIMIT $2 OFFSET $3`

	rows, err := r.getConn(ctx).Query(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.FromWalletID,
			&t.ToWalletID,
			&t.Amount,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, &t)
	}

	return transactions, nil
}

func (r *PostgresRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.connPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("tx error: %w", err)
	}

	ctx = context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx rollback error: %w", rbErr)
		}
		return fmt.Errorf("tx error: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx error: %w", err)
	}

	return nil
}

func (r *PostgresRepo) CreateTransaction(ctx context.Context, transaction domain.Transaction) error {
	query := `INSERT INTO transactions(id, from_wallet_id, to_wallet_id, amount) VALUES($1, $2, $3, $4)`

	if _, err := r.getConn(ctx).Exec(
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
