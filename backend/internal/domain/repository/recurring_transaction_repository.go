package repository

import (
	"context"
	"time"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type RecurringTransactionRepository interface {
	Create(ctx context.Context, rt *entity.RecurringTransaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.RecurringTransaction, error)
	FindAll(ctx context.Context, userID uuid.UUID, filter entity.RecurringTransactionFilter) (*entity.PaginatedRecurringTransactions, error)
	Pause(ctx context.Context, id uuid.UUID, pausedAt time.Time) error
	Resume(ctx context.Context, id uuid.UUID) error
}
