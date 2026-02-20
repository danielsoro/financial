package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/google/uuid"
)

type DashboardUsecase struct {
	transactionRepo  repository.TransactionRepository
	expenseLimitRepo repository.ExpenseLimitRepository
}

func NewDashboardUsecase(
	transactionRepo repository.TransactionRepository,
	expenseLimitRepo repository.ExpenseLimitRepository,
) *DashboardUsecase {
	return &DashboardUsecase{
		transactionRepo:  transactionRepo,
		expenseLimitRepo: expenseLimitRepo,
	}
}

func (uc *DashboardUsecase) GetSummary(ctx context.Context, month, year int, userID *uuid.UUID) (*entity.DashboardSummary, error) {
	return uc.transactionRepo.GetSummary(ctx, month, year, userID)
}

func (uc *DashboardUsecase) GetByCategory(ctx context.Context, month, year int, txType string, userID *uuid.UUID) ([]entity.CategoryTotal, error) {
	return uc.transactionRepo.GetByCategory(ctx, month, year, txType, userID)
}

func (uc *DashboardUsecase) GetLimitsProgress(ctx context.Context, month, year int, userID *uuid.UUID) ([]entity.LimitProgress, error) {
	return uc.expenseLimitRepo.GetLimitsProgress(ctx, month, year, userID)
}
