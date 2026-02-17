package database

import (
	"context"
	"errors"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		user.Name, user.Email, user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var u entity.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, user *entity.User) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2, password_hash = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING updated_at`,
		user.Name, user.Email, user.PasswordHash, user.ID,
	).Scan(&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if isDuplicateKey(err) {
			return domain.ErrDuplicateEmail
		}
		return err
	}
	return nil
}
