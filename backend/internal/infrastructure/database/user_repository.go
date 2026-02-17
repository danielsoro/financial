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
		`INSERT INTO users (tenant_id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		user.TenantID, user.Name, user.Email, user.PasswordHash, user.Role,
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
		`SELECT id, tenant_id, name, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.TenantID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
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
		`SELECT id, tenant_id, name, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.TenantID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
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
		`UPDATE users SET name = $1, email = $2, password_hash = $3, role = $4, updated_at = NOW()
		 WHERE id = $5
		 RETURNING updated_at`,
		user.Name, user.Email, user.PasswordHash, user.Role, user.ID,
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

func (r *UserRepo) FindAllByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AdminUser, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, email, role, created_at, updated_at
		 FROM users WHERE tenant_id = $1 ORDER BY created_at ASC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.AdminUser
	for rows.Next() {
		var u entity.AdminUser
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if users == nil {
		users = []entity.AdminUser{}
	}
	return users, nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
