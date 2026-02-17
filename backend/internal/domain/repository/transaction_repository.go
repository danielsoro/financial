package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	Update(ctx context.Context, tx *entity.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	FindAll(ctx context.Context, filter entity.TransactionFilter) (*entity.PaginatedTransactions, error)
	GetSummary(ctx context.Context, userID uuid.UUID, month, year int) (*entity.DashboardSummary, error)
	GetByCategory(ctx context.Context, userID uuid.UUID, month, year int, txType string) ([]entity.CategoryTotal, error)
}
