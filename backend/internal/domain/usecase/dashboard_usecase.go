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

func (uc *DashboardUsecase) GetSummary(ctx context.Context, tenantID uuid.UUID, month, year int) (*entity.DashboardSummary, error) {
	return uc.transactionRepo.GetSummary(ctx, month, year)
}

func (uc *DashboardUsecase) GetByCategory(ctx context.Context, tenantID uuid.UUID, month, year int, txType string) ([]entity.CategoryTotal, error) {
	return uc.transactionRepo.GetByCategory(ctx, month, year, txType)
}

func (uc *DashboardUsecase) GetLimitsProgress(ctx context.Context, tenantID uuid.UUID, month, year int) ([]entity.LimitProgress, error) {
	return uc.expenseLimitRepo.GetLimitsProgress(ctx, month, year)
}
