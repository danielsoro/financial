package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepo struct {
	pool *pgxpool.Pool
}

func NewTenantRepo(pool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{pool: pool}
}

func (r *TenantRepo) Create(ctx context.Context, tenant *entity.Tenant) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tenants (name, domain, schema_name, is_active) VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		tenant.Name, tenant.Domain, tenant.SchemaName, tenant.IsActive,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateDomain
		}
		return err
	}
	return nil
}

func (r *TenantRepo) Update(ctx context.Context, tenant *entity.Tenant) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE tenants SET name = $1, domain = $2, schema_name = $3, is_active = $4, updated_at = NOW()
		 WHERE id = $5
		 RETURNING updated_at`,
		tenant.Name, tenant.Domain, tenant.SchemaName, tenant.IsActive, tenant.ID,
	).Scan(&tenant.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if isDuplicateKey(err) {
			return domain.ErrDuplicateDomain
		}
		return err
	}
	return nil
}

func (r *TenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if t.SchemaName != "" {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, pgx.Identifier{t.SchemaName}.Sanitize())); err != nil {
			return err
		}
	}

	result, err := tx.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *TenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var t entity.Tenant
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, domain, schema_name, is_active, created_at, updated_at FROM tenants WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Domain, &t.SchemaName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) FindByDomain(ctx context.Context, domainStr string) (*entity.Tenant, error) {
	var t entity.Tenant
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, domain, schema_name, is_active, created_at, updated_at FROM tenants WHERE domain = $1`, domainStr,
	).Scan(&t.ID, &t.Name, &t.Domain, &t.SchemaName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) FindAll(ctx context.Context) ([]entity.Tenant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, domain, schema_name, is_active, created_at, updated_at FROM tenants ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []entity.Tenant
	for rows.Next() {
		var t entity.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Domain, &t.SchemaName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tenants == nil {
		tenants = []entity.Tenant{}
	}
	return tenants, nil
}
