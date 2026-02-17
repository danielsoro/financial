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

type CategoryRepo struct {
	pool *pgxpool.Pool
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

func (r *CategoryRepo) Create(ctx context.Context, cat *entity.Category) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (tenant_id, user_id, parent_id, name, type, is_default)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		cat.TenantID, cat.UserID, cat.ParentID, cat.Name, cat.Type, cat.IsDefault,
	).Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateCategory
		}
		return err
	}
	return nil
}

func (r *CategoryRepo) Update(ctx context.Context, cat *entity.Category) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE categories SET name = $1, type = $2, parent_id = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING updated_at`,
		cat.Name, cat.Type, cat.ParentID, cat.ID,
	).Scan(&cat.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if isDuplicateKey(err) {
			return domain.ErrDuplicateCategory
		}
		return err
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var cat entity.Category
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, parent_id, name, type, is_default, created_at, updated_at
		 FROM categories WHERE id = $1`, id,
	).Scan(&cat.ID, &cat.TenantID, &cat.UserID, &cat.ParentID, &cat.Name, &cat.Type, &cat.IsDefault, &cat.CreatedAt, &cat.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) FindAllForTenant(ctx context.Context, tenantID uuid.UUID, catType string) ([]entity.Category, error) {
	query := `WITH RECURSIVE cat_tree AS (
		SELECT id, tenant_id, user_id, parent_id, name, type, is_default, created_at, updated_at,
		       name::text AS full_path
		FROM categories
		WHERE parent_id IS NULL AND tenant_id = $1
		UNION ALL
		SELECT c.id, c.tenant_id, c.user_id, c.parent_id, c.name, c.type, c.is_default, c.created_at, c.updated_at,
		       ct.full_path || ' > ' || c.name
		FROM categories c
		INNER JOIN cat_tree ct ON c.parent_id = ct.id
	)
	SELECT id, tenant_id, user_id, parent_id, name, type, is_default, created_at, updated_at, full_path
	FROM cat_tree
	WHERE 1=1`

	args := []any{tenantID}
	argIdx := 2

	if catType != "" {
		if catType == "income" {
			query += fmt.Sprintf(` AND type IN ($%d, 'both')`, argIdx)
			args = append(args, "income")
			argIdx++
		} else if catType == "expense" {
			query += fmt.Sprintf(` AND type IN ($%d, 'both')`, argIdx)
			args = append(args, "expense")
			argIdx++
		}
	}
	_ = argIdx

	query += ` ORDER BY full_path ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var cat entity.Category
		if err := rows.Scan(&cat.ID, &cat.TenantID, &cat.UserID, &cat.ParentID, &cat.Name, &cat.Type, &cat.IsDefault, &cat.CreatedAt, &cat.UpdatedAt, &cat.FullPath); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepo) IsInUse(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM transactions WHERE category_id = $1)`, id,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *CategoryRepo) IsSubtreeInUse(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`WITH RECURSIVE subtree AS (
			SELECT id FROM categories WHERE id = $1
			UNION ALL
			SELECT c.id FROM categories c INNER JOIN subtree s ON c.parent_id = s.id
		)
		SELECT EXISTS(SELECT 1 FROM transactions WHERE category_id IN (SELECT id FROM subtree))`, id,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
