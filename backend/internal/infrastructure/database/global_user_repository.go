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

type GlobalUserRepo struct {
	pool *pgxpool.Pool
}

func NewGlobalUserRepo(pool *pgxpool.Pool) *GlobalUserRepo {
	return &GlobalUserRepo{pool: pool}
}

func (r *GlobalUserRepo) Create(ctx context.Context, user *entity.GlobalUser) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO global_users (name, email, password_hash, email_verified, email_token, email_token_expires_at, max_owned_tenants)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		user.Name, user.Email, user.PasswordHash, user.EmailVerified, user.EmailToken, user.EmailTokenExpiresAt, user.MaxOwnedTenants,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *GlobalUserRepo) FindByEmail(ctx context.Context, email string) (*entity.GlobalUser, error) {
	var u entity.GlobalUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, email_verified, email_token, email_token_expires_at, max_owned_tenants, created_at, updated_at
		 FROM global_users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.EmailToken, &u.EmailTokenExpiresAt, &u.MaxOwnedTenants, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GlobalUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.GlobalUser, error) {
	var u entity.GlobalUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, email_verified, email_token, email_token_expires_at, max_owned_tenants, created_at, updated_at
		 FROM global_users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.EmailToken, &u.EmailTokenExpiresAt, &u.MaxOwnedTenants, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GlobalUserRepo) Update(ctx context.Context, user *entity.GlobalUser) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE global_users SET name = $1, email = $2, password_hash = $3, email_verified = $4,
		 email_token = $5, email_token_expires_at = $6, max_owned_tenants = $7, updated_at = NOW()
		 WHERE id = $8
		 RETURNING updated_at`,
		user.Name, user.Email, user.PasswordHash, user.EmailVerified,
		user.EmailToken, user.EmailTokenExpiresAt, user.MaxOwnedTenants, user.ID,
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

func (r *GlobalUserRepo) FindByEmailToken(ctx context.Context, token string) (*entity.GlobalUser, error) {
	var u entity.GlobalUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, email_verified, email_token, email_token_expires_at, max_owned_tenants, created_at, updated_at
		 FROM global_users WHERE email_token = $1`, token,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.EmailToken, &u.EmailTokenExpiresAt, &u.MaxOwnedTenants, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GlobalUserRepo) CountOwnedTenants(ctx context.Context, globalUserID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM memberships WHERE global_user_id = $1 AND role = 'owner'`, globalUserID,
	).Scan(&count)
	return count, err
}
