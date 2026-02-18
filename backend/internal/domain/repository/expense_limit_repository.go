package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type ExpenseLimitRepository interface {
	Upsert(ctx context.Context, limit *entity.ExpenseLimit) error
	Update(ctx context.Context, limit *entity.ExpenseLimit) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ExpenseLimit, error)
	FindAll(ctx context.Context, month, year int) ([]entity.ExpenseLimit, error)
	GetLimitsProgress(ctx context.Context, month, year int) ([]entity.LimitProgress, error)
}
