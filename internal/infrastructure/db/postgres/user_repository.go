package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		u.ID,
		u.Email.String(),
		string(u.PasswordHash),
		u.FirstName,
		u.LastName,
		u.CreatedAt,
		u.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4, updated_at = $5
		WHERE id = $6`

	commandTag, err := r.pool.Exec(ctx, query,
		u.Email.String(),
		string(u.PasswordHash),
		u.FirstName,
		u.LastName,
		u.UpdatedAt,
		u.ID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrUserAlreadyExists
		}
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users WHERE id = $1`

	var u user.User
	var rawEmail, rawHash string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&rawEmail,
		&rawHash,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	parsedEmail, _ := user.NewEmail(rawEmail)
	u.Email = parsedEmail
	u.PasswordHash = user.HashedPassword(rawHash)

	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users WHERE email = $1`

	var u user.User
	var rawEmail, rawHash string

	err := r.pool.QueryRow(ctx, query, email.String()).Scan(
		&u.ID,
		&rawEmail,
		&rawHash,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	parsedEmail, _ := user.NewEmail(rawEmail)
	u.Email = parsedEmail
	u.PasswordHash = user.HashedPassword(rawHash)

	return &u, nil
}
