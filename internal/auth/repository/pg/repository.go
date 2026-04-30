package pg_repo

import (
	"context"
	"errors"

	"github.com/AriartyyyA/gobank/internal/auth/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	connPool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) domain.UserRepository {
	return &PostgresRepo{
		connPool: pool,
	}
}

func (r *PostgresRepo) CreateUser(ctx context.Context, user domain.User) error {
	query := `INSERT INTO users(id, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`

	if _, err := r.connPool.Exec(
		ctx,
		query,
		user.UUID,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserExists
		}

		return err
	}

	return nil
}

func (r *PostgresRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`

	row := r.connPool.QueryRow(ctx, query, email)

	var user domain.User
	if err := row.Scan(
		&user.UUID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1`

	row := r.connPool.QueryRow(ctx, query, id)

	var user domain.User
	if err := row.Scan(
		&user.UUID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}
