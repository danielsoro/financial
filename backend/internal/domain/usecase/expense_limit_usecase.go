package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/google/uuid"
)

type ExpenseLimitUsecase struct {
	expenseLimitRepo repository.ExpenseLimitRepository
}

func NewExpenseLimitUsecase(repo repository.ExpenseLimitRepository) *ExpenseLimitUsecase {
	return &ExpenseLimitUsecase{expenseLimitRepo: repo}
}

func (uc *ExpenseLimitUsecase) List(ctx context.Context, tenantID uuid.UUID, month, year int) ([]entity.ExpenseLimit, error) {
	return uc.expenseLimitRepo.FindAll(ctx, month, year)
}

func (uc *ExpenseLimitUsecase) Create(ctx context.Context, limit *entity.ExpenseLimit) error {
	return uc.expenseLimitRepo.Upsert(ctx, limit)
}

func (uc *ExpenseLimitUsecase) Update(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, amount float64) error {
	limit, err := uc.expenseLimitRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	limit.Amount = amount
	return uc.expenseLimitRepo.Update(ctx, limit)
}

func (uc *ExpenseLimitUsecase) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	if _, err := uc.expenseLimitRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.expenseLimitRepo.Delete(ctx, id)
}
