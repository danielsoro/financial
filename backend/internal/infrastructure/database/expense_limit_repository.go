package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ExpenseLimitRepo struct{}

func NewExpenseLimitRepo() *ExpenseLimitRepo {
	return &ExpenseLimitRepo{}
}

func (r *ExpenseLimitRepo) Upsert(ctx context.Context, limit *entity.ExpenseLimit) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	if limit.CategoryID != nil {
		err := conn.QueryRow(ctx,
			`INSERT INTO expense_limits (user_id, category_id, month, year, amount)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (category_id, month, year)
			 DO UPDATE SET amount = EXCLUDED.amount, updated_at = NOW()
			 RETURNING id, created_at, updated_at`,
			limit.UserID, limit.CategoryID, limit.Month, limit.Year, limit.Amount,
		).Scan(&limit.ID, &limit.CreatedAt, &limit.UpdatedAt)
		if err != nil {
			return err
		}
		return nil
	}

	err = conn.QueryRow(ctx,
		`WITH existing AS (
			SELECT id FROM expense_limits
			WHERE category_id IS NULL AND month = $1 AND year = $2
		),
		updated AS (
			UPDATE expense_limits
			SET amount = $3, updated_at = NOW()
			WHERE id = (SELECT id FROM existing)
			RETURNING id, created_at, updated_at
		),
		inserted AS (
			INSERT INTO expense_limits (user_id, category_id, month, year, amount)
			SELECT $4, NULL, $1, $2, $3
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING id, created_at, updated_at
		)
		SELECT id, created_at, updated_at FROM updated
		UNION ALL
		SELECT id, created_at, updated_at FROM inserted`,
		limit.Month, limit.Year, limit.Amount, limit.UserID,
	).Scan(&limit.ID, &limit.CreatedAt, &limit.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *ExpenseLimitRepo) Update(ctx context.Context, limit *entity.ExpenseLimit) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryRow(ctx,
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
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	result, err := conn.Exec(ctx, `DELETE FROM expense_limits WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ExpenseLimitRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.ExpenseLimit, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var limit entity.ExpenseLimit
	var categoryName *string
	err = conn.QueryRow(ctx,
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

func (r *ExpenseLimitRepo) FindAll(ctx context.Context, month, year int) ([]entity.ExpenseLimit, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx,
		`SELECT el.id, el.user_id, el.category_id, c.name AS category_name,
		        el.month, el.year, el.amount, el.created_at, el.updated_at
		 FROM expense_limits el
		 LEFT JOIN categories c ON el.category_id = c.id
		 WHERE el.month = $1 AND el.year = $2
		 ORDER BY el.created_at ASC`,
		month, year,
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

func (r *ExpenseLimitRepo) GetLimitsProgress(ctx context.Context, month, year int, userID *uuid.UUID) ([]entity.LimitProgress, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userFilterOuter := ""
	userFilterLateral := ""
	args := []any{month, year}
	if userID != nil {
		argIdx := len(args) + 1
		userFilterOuter = fmt.Sprintf(" AND el.user_id = $%d", argIdx)
		userFilterLateral = fmt.Sprintf(" AND t.user_id = $%d", argIdx)
		args = append(args, *userID)
	}

	query := fmt.Sprintf(
		`SELECT el.id, el.user_id, el.category_id, c.name AS category_name,
		        el.month, el.year, el.amount, el.created_at, el.updated_at,
		        COALESCE(spent.total, 0) AS spent
		 FROM expense_limits el
		 LEFT JOIN categories c ON el.category_id = c.id
		 LEFT JOIN LATERAL (
			SELECT SUM(t.amount) AS total
			FROM transactions t
			WHERE t.type = 'expense'
			  AND EXTRACT(MONTH FROM t.date::date) = el.month
			  AND EXTRACT(YEAR FROM t.date::date) = el.year
			  AND (
				  el.category_id IS NULL
				  OR t.category_id = el.category_id
			  )
			  %s
		 ) spent ON true
		 WHERE el.month = $1 AND el.year = $2
		 %s
		 ORDER BY el.created_at ASC`, userFilterLateral, userFilterOuter)

	rows, err := conn.Query(ctx, query, args...)
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
