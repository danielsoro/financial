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

type ExpenseLimitRepo struct {
	pool *pgxpool.Pool
}

func NewExpenseLimitRepo(pool *pgxpool.Pool) *ExpenseLimitRepo {
	return &ExpenseLimitRepo{pool: pool}
}

func (r *ExpenseLimitRepo) Upsert(ctx context.Context, limit *entity.ExpenseLimit) error {
	if limit.CategoryID != nil {
		// Category-specific limit: use ON CONFLICT with the unique index
		err := r.pool.QueryRow(ctx,
			`INSERT INTO expense_limits (user_id, category_id, month, year, amount)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (user_id, category_id, month, year)
			 DO UPDATE SET amount = EXCLUDED.amount, updated_at = NOW()
			 RETURNING id, created_at, updated_at`,
			limit.UserID, limit.CategoryID, limit.Month, limit.Year, limit.Amount,
		).Scan(&limit.ID, &limit.CreatedAt, &limit.UpdatedAt)
		if err != nil {
			return err
		}
		return nil
	}

	// Global limit (category_id IS NULL): use a CTE to handle the partial unique index
	err := r.pool.QueryRow(ctx,
		`WITH existing AS (
			SELECT id FROM expense_limits
			WHERE user_id = $1 AND category_id IS NULL AND month = $2 AND year = $3
		),
		updated AS (
			UPDATE expense_limits
			SET amount = $4, updated_at = NOW()
			WHERE id = (SELECT id FROM existing)
			RETURNING id, created_at, updated_at
		),
		inserted AS (
			INSERT INTO expense_limits (user_id, category_id, month, year, amount)
			SELECT $1, NULL, $2, $3, $4
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING id, created_at, updated_at
		)
		SELECT id, created_at, updated_at FROM updated
		UNION ALL
		SELECT id, created_at, updated_at FROM inserted`,
		limit.UserID, limit.Month, limit.Year, limit.Amount,
	).Scan(&limit.ID, &limit.CreatedAt, &limit.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *ExpenseLimitRepo) Update(ctx context.Context, limit *entity.ExpenseLimit) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE expense_limits SET amount = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING updated_at`,
		limit.Amount, limit.ID,
	).Scan(&limit.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *ExpenseLimitRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM expense_limits WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ExpenseLimitRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.ExpenseLimit, error) {
	var limit entity.ExpenseLimit
	var categoryName *string
	err := r.pool.QueryRow(ctx,
		`SELECT el.id, el.user_id, el.category_id, c.name AS category_name,
		        el.month, el.year, el.amount, el.created_at, el.updated_at
		 FROM expense_limits el
		 LEFT JOIN categories c ON el.category_id = c.id
		 WHERE el.id = $1`, id,
	).Scan(&limit.ID, &limit.UserID, &limit.CategoryID, &categoryName,
		&limit.Month, &limit.Year, &limit.Amount, &limit.CreatedAt, &limit.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if categoryName != nil {
		limit.CategoryName = *categoryName
	}
	return &limit, nil
}

func (r *ExpenseLimitRepo) FindAll(ctx context.Context, userID uuid.UUID, month, year int) ([]entity.ExpenseLimit, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT el.id, el.user_id, el.category_id, c.name AS category_name,
		        el.month, el.year, el.amount, el.created_at, el.updated_at
		 FROM expense_limits el
		 LEFT JOIN categories c ON el.category_id = c.id
		 WHERE el.user_id = $1 AND el.month = $2 AND el.year = $3
		 ORDER BY el.created_at ASC`,
		userID, month, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []entity.ExpenseLimit
	for rows.Next() {
		var limit entity.ExpenseLimit
		var categoryName *string
		if err := rows.Scan(&limit.ID, &limit.UserID, &limit.CategoryID, &categoryName,
			&limit.Month, &limit.Year, &limit.Amount, &limit.CreatedAt, &limit.UpdatedAt); err != nil {
			return nil, err
		}
		if categoryName != nil {
			limit.CategoryName = *categoryName
		}
		limits = append(limits, limit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if limits == nil {
		limits = []entity.ExpenseLimit{}
	}

	return limits, nil
}

func (r *ExpenseLimitRepo) GetLimitsProgress(ctx context.Context, userID uuid.UUID, month, year int) ([]entity.LimitProgress, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT el.id, el.user_id, el.category_id, c.name AS category_name,
		        el.month, el.year, el.amount, el.created_at, el.updated_at,
		        COALESCE(spent.total, 0) AS spent
		 FROM expense_limits el
		 LEFT JOIN categories c ON el.category_id = c.id
		 LEFT JOIN LATERAL (
			SELECT SUM(t.amount) AS total
			FROM transactions t
			WHERE t.user_id = el.user_id
			  AND t.type = 'expense'
			  AND EXTRACT(MONTH FROM t.date::date) = el.month
			  AND EXTRACT(YEAR FROM t.date::date) = el.year
			  AND (
				  el.category_id IS NULL
				  OR t.category_id = el.category_id
			  )
		 ) spent ON true
		 WHERE el.user_id = $1 AND el.month = $2 AND el.year = $3
		 ORDER BY el.created_at ASC`,
		userID, month, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []entity.LimitProgress
	for rows.Next() {
		var lp entity.LimitProgress
		var categoryName *string
		if err := rows.Scan(
			&lp.Limit.ID, &lp.Limit.UserID, &lp.Limit.CategoryID, &categoryName,
			&lp.Limit.Month, &lp.Limit.Year, &lp.Limit.Amount,
			&lp.Limit.CreatedAt, &lp.Limit.UpdatedAt,
			&lp.Spent,
		); err != nil {
			return nil, err
		}
		if categoryName != nil {
			lp.Limit.CategoryName = *categoryName
		}
		lp.Remaining = lp.Limit.Amount - lp.Spent
		if lp.Remaining < 0 {
			lp.Remaining = 0
		}
		if lp.Limit.Amount > 0 {
			lp.Percentage = (lp.Spent / lp.Limit.Amount) * 100
		}
		progress = append(progress, lp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if progress == nil {
		progress = []entity.LimitProgress{}
	}

	return progress, nil
}
