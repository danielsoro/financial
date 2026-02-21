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

type TransactionRepo struct{}

func NewTransactionRepo() *TransactionRepo {
	return &TransactionRepo{}
}

func (r *TransactionRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryRow(ctx,
		`INSERT INTO transactions (user_id, category_id, type, amount, description, date, recurring_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		tx.UserID, tx.CategoryID, tx.Type, tx.Amount, tx.Description, tx.Date, tx.RecurringID,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *TransactionRepo) BulkCreate(ctx context.Context, txs []entity.Transaction) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for i := range txs {
		batch.Queue(
			`INSERT INTO transactions (user_id, category_id, type, amount, description, date, recurring_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			txs[i].UserID, txs[i].CategoryID, txs[i].Type, txs[i].Amount, txs[i].Description, txs[i].Date, txs[i].RecurringID,
		)
	}

	br := conn.SendBatch(ctx, batch)
	defer br.Close()

	for range txs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *TransactionRepo) DeleteByRecurringID(ctx context.Context, recurringID uuid.UUID, mode entity.DeleteMode) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	var query string

	switch mode {
	case entity.DeleteModeAll:
		query = `DELETE FROM transactions WHERE recurring_id = $1`
		_, err = conn.Exec(ctx, query, recurringID)
	case entity.DeleteModeFutureAndCurrent:
		firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		query = `DELETE FROM transactions WHERE recurring_id = $1 AND date >= $2`
		_, err = conn.Exec(ctx, query, recurringID, firstDayOfMonth)
	case entity.DeleteModeFutureOnly:
		firstDayNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		query = `DELETE FROM transactions WHERE recurring_id = $1 AND date >= $2`
		_, err = conn.Exec(ctx, query, recurringID, firstDayNextMonth)
	}
	return err
}

func (r *TransactionRepo) DeleteFutureByRecurringID(ctx context.Context, recurringID uuid.UUID, fromDate string) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx,
		`DELETE FROM transactions WHERE recurring_id = $1 AND date >= $2`,
		recurringID, fromDate)
	return err
}

func (r *TransactionRepo) CountByRecurringID(ctx context.Context, recurringID uuid.UUID) (int, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE recurring_id = $1`, recurringID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TransactionRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryRow(ctx,
		`UPDATE transactions
		 SET type = $1, amount = $2, description = $3, date = $4, category_id = $5, updated_at = NOW()
		 WHERE id = $6
		 RETURNING updated_at`,
		tx.Type, tx.Amount, tx.Description, tx.Date, tx.CategoryID, tx.ID,
	).Scan(&tx.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *TransactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return err
	}

	result, err := conn.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var tx entity.Transaction
	err = conn.QueryRow(ctx,
		`SELECT t.id, t.user_id, t.category_id, c.name AS category_name,
		        t.type, t.amount, t.description, t.date::text, t.recurring_id, t.created_at, t.updated_at
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.id = $1`, id,
	).Scan(&tx.ID, &tx.UserID, &tx.CategoryID, &tx.CategoryName,
		&tx.Type, &tx.Amount, &tx.Description, &tx.Date, &tx.RecurringID, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepo) FindAll(ctx context.Context, filter entity.TransactionFilter) (*entity.PaginatedTransactions, error) {
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

	baseWhere := ` WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Type != "" {
		baseWhere += fmt.Sprintf(` AND t.type = $%d`, argIdx)
		args = append(args, filter.Type)
		argIdx++
	}

	if filter.CategoryID != nil {
		baseWhere += fmt.Sprintf(` AND t.category_id = $%d`, argIdx)
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	if filter.StartDate != "" {
		baseWhere += fmt.Sprintf(` AND t.date >= $%d`, argIdx)
		args = append(args, filter.StartDate)
		argIdx++
	}

	if filter.EndDate != "" {
		baseWhere += fmt.Sprintf(` AND t.date <= $%d`, argIdx)
		args = append(args, filter.EndDate)
		argIdx++
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transactions t%s`, baseWhere)
	var total int
	if err := conn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PerPage)))

	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT t.id, t.user_id, t.category_id, c.name AS category_name,
		        t.type, t.amount, t.description, t.date::text, t.recurring_id, t.created_at, t.updated_at
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 %s
		 ORDER BY t.date DESC, t.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		baseWhere, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []entity.Transaction
	for rows.Next() {
		var tx entity.Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.CategoryID, &tx.CategoryName,
			&tx.Type, &tx.Amount, &tx.Description, &tx.Date, &tx.RecurringID, &tx.CreatedAt, &tx.UpdatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []entity.Transaction{}
	}

	return &entity.PaginatedTransactions{
		Data:       transactions,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (r *TransactionRepo) GetSummary(ctx context.Context, month, year int, userID *uuid.UUID) (*entity.DashboardSummary, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userFilter := ""
	args := []any{month, year}
	if userID != nil {
		userFilter = fmt.Sprintf(" AND user_id = $%d", len(args)+1)
		args = append(args, *userID)
	}

	query := fmt.Sprintf(`
		WITH current_month AS (
			SELECT
				COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income,
				COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expenses,
				COUNT(*) FILTER (WHERE type = 'income') AS income_count,
				COUNT(*) FILTER (WHERE type = 'expense') AS expense_count
			FROM transactions
			WHERE EXTRACT(MONTH FROM date::date) = $1
			  AND EXTRACT(YEAR FROM date::date) = $2
			  %s
		),
		previous_months AS (
			SELECT
				COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0) AS balance
			FROM transactions
			WHERE date < make_date($2::int, $1::int, 1)
			  %s
		)
		SELECT cm.income, cm.expenses, cm.income_count, cm.expense_count, pm.balance
		FROM current_month cm, previous_months pm`, userFilter, userFilter)

	summary := &entity.DashboardSummary{}
	err = conn.QueryRow(ctx, query, args...).Scan(
		&summary.TotalIncome, &summary.TotalExpenses,
		&summary.IncomeCount, &summary.ExpenseCount,
		&summary.PreviousBalance,
	)
	if err != nil {
		return nil, err
	}

	summary.Balance = summary.PreviousBalance + summary.TotalIncome - summary.TotalExpenses

	return summary, nil
}

func (r *TransactionRepo) GetByCategory(ctx context.Context, month, year int, txType string, userID *uuid.UUID) ([]entity.CategoryTotal, error) {
	conn, err := ConnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userFilter := ""
	args := []any{month, year, txType}
	if userID != nil {
		userFilter = fmt.Sprintf(" AND t.user_id = $%d", len(args)+1)
		args = append(args, *userID)
	}

	query := fmt.Sprintf(
		`SELECT t.category_id, c.name AS category_name, SUM(t.amount) AS total
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE EXTRACT(MONTH FROM t.date::date) = $1
		   AND EXTRACT(YEAR FROM t.date::date) = $2
		   AND t.type = $3
		   %s
		 GROUP BY t.category_id, c.name
		 ORDER BY total DESC`, userFilter)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []entity.CategoryTotal
	for rows.Next() {
		var ct entity.CategoryTotal
		if err := rows.Scan(&ct.CategoryID, &ct.CategoryName, &ct.Total); err != nil {
			return nil, err
		}
		totals = append(totals, ct)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if totals == nil {
		totals = []entity.CategoryTotal{}
	}

	return totals, nil
}
