package database

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

func (r *TransactionRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO transactions (user_id, category_id, type, amount, description, date)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		tx.UserID, tx.CategoryID, tx.Type, tx.Amount, tx.Description, tx.Date,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *TransactionRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	err := r.pool.QueryRow(ctx,
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
	result, err := r.pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var tx entity.Transaction
	err := r.pool.QueryRow(ctx,
		`SELECT t.id, t.user_id, t.category_id, c.name AS category_name,
		        t.type, t.amount, t.description, t.date::text, t.created_at, t.updated_at
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.id = $1`, id,
	).Scan(&tx.ID, &tx.UserID, &tx.CategoryID, &tx.CategoryName,
		&tx.Type, &tx.Amount, &tx.Description, &tx.Date, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepo) FindAll(ctx context.Context, filter entity.TransactionFilter) (*entity.PaginatedTransactions, error) {
	if filter.PerPage <= 0 {
		filter.PerPage = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	baseWhere := ` WHERE t.user_id = $1`
	args := []any{filter.UserID}
	argIdx := 2

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

	// Count total matching rows
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transactions t%s`, baseWhere)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PerPage)))

	// Fetch paginated results
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT t.id, t.user_id, t.category_id, c.name AS category_name,
		        t.type, t.amount, t.description, t.date::text, t.created_at, t.updated_at
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 %s
		 ORDER BY t.date DESC, t.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		baseWhere, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []entity.Transaction
	for rows.Next() {
		var tx entity.Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.CategoryID, &tx.CategoryName,
			&tx.Type, &tx.Amount, &tx.Description, &tx.Date, &tx.CreatedAt, &tx.UpdatedAt); err != nil {
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

func (r *TransactionRepo) GetSummary(ctx context.Context, userID uuid.UUID, month, year int) (*entity.DashboardSummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT type,
		        COALESCE(SUM(amount), 0) AS total,
		        COUNT(*) AS count
		 FROM transactions
		 WHERE user_id = $1
		   AND EXTRACT(MONTH FROM date::date) = $2
		   AND EXTRACT(YEAR FROM date::date) = $3
		 GROUP BY type`,
		userID, month, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := &entity.DashboardSummary{}
	for rows.Next() {
		var txType string
		var total float64
		var count int
		if err := rows.Scan(&txType, &total, &count); err != nil {
			return nil, err
		}
		switch txType {
		case "income":
			summary.TotalIncome = total
			summary.IncomeCount = count
		case "expense":
			summary.TotalExpenses = total
			summary.ExpenseCount = count
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	summary.Balance = summary.TotalIncome - summary.TotalExpenses

	return summary, nil
}

func (r *TransactionRepo) GetByCategory(ctx context.Context, userID uuid.UUID, month, year int, txType string) ([]entity.CategoryTotal, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.category_id, c.name AS category_name, SUM(t.amount) AS total
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = $1
		   AND EXTRACT(MONTH FROM t.date::date) = $2
		   AND EXTRACT(YEAR FROM t.date::date) = $3
		   AND t.type = $4
		 GROUP BY t.category_id, c.name
		 ORDER BY total DESC`,
		userID, month, year, txType,
	)
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
