package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RecurringTransactionRepo struct{}

func NewRecurringTransactionRepo() *RecurringTransactionRepo {
	return &RecurringTransactionRepo{}
}

func (r *RecurringTransactionRepo) Create(ctx context.Context, rt *entity.RecurringTransaction) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryRow(ctx,
		`INSERT INTO recurring_transactions (user_id, category_id, type, amount, description, frequency, start_date, end_date, max_occurrences, day_of_month, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, created_at, updated_at`,
		rt.UserID, rt.CategoryID, rt.Type, rt.Amount, rt.Description, rt.Frequency,
		rt.StartDate, rt.EndDate, rt.MaxOccurrences, rt.DayOfMonth, rt.IsActive,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *RecurringTransactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	result, err := conn.Exec(ctx, `DELETE FROM recurring_transactions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RecurringTransactionRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.RecurringTransaction, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var rt entity.RecurringTransaction
	err = conn.QueryRow(ctx,
		`SELECT rt.id, rt.user_id, rt.category_id, c.name AS category_name,
		        rt.type, rt.amount, rt.description, rt.frequency,
		        rt.start_date::text, rt.end_date::text, rt.max_occurrences, rt.day_of_month,
		        rt.is_active, rt.paused_at, rt.created_at, rt.updated_at
		 FROM recurring_transactions rt
		 JOIN categories c ON rt.category_id = c.id
		 WHERE rt.id = $1`, id,
	).Scan(&rt.ID, &rt.UserID, &rt.CategoryID, &rt.CategoryName,
		&rt.Type, &rt.Amount, &rt.Description, &rt.Frequency,
		&rt.StartDate, &rt.EndDate, &rt.MaxOccurrences, &rt.DayOfMonth,
		&rt.IsActive, &rt.PausedAt, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func (r *RecurringTransactionRepo) FindAll(ctx context.Context, userID uuid.UUID, filter entity.RecurringTransactionFilter) (*entity.PaginatedRecurringTransactions, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if filter.PerPage <= 0 {
		filter.PerPage = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	baseWhere := ` WHERE rt.user_id = $1`
	args := []any{userID}
	argIdx := 2

	if filter.Type != "" {
		baseWhere += fmt.Sprintf(` AND rt.type = $%d`, argIdx)
		args = append(args, filter.Type)
		argIdx++
	}

	if filter.IsActive != nil {
		baseWhere += fmt.Sprintf(` AND rt.is_active = $%d`, argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM recurring_transactions rt%s`, baseWhere)
	var total int
	if err := conn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PerPage)))

	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT rt.id, rt.user_id, rt.category_id, c.name AS category_name,
		        rt.type, rt.amount, rt.description, rt.frequency,
		        rt.start_date::text, rt.end_date::text, rt.max_occurrences, rt.day_of_month,
		        rt.is_active, rt.paused_at, rt.created_at, rt.updated_at
		 FROM recurring_transactions rt
		 JOIN categories c ON rt.category_id = c.id
		 %s
		 ORDER BY rt.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		baseWhere, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.RecurringTransaction
	for rows.Next() {
		var rt entity.RecurringTransaction
		if err := rows.Scan(&rt.ID, &rt.UserID, &rt.CategoryID, &rt.CategoryName,
			&rt.Type, &rt.Amount, &rt.Description, &rt.Frequency,
			&rt.StartDate, &rt.EndDate, &rt.MaxOccurrences, &rt.DayOfMonth,
			&rt.IsActive, &rt.PausedAt, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []entity.RecurringTransaction{}
	}

	return &entity.PaginatedRecurringTransactions{
		Data:       items,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (r *RecurringTransactionRepo) Pause(ctx context.Context, id uuid.UUID, pausedAt time.Time) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	result, err := conn.Exec(ctx,
		`UPDATE recurring_transactions SET is_active = false, paused_at = $1, updated_at = NOW() WHERE id = $2`,
		pausedAt, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RecurringTransactionRepo) Resume(ctx context.Context, id uuid.UUID) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	result, err := conn.Exec(ctx,
		`UPDATE recurring_transactions SET is_active = true, paused_at = NULL, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
