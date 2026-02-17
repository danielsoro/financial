package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain"
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

func (uc *ExpenseLimitUsecase) List(ctx context.Context, userID uuid.UUID, month, year int) ([]entity.ExpenseLimit, error) {
	return uc.expenseLimitRepo.FindAll(ctx, userID, month, year)
}

func (uc *ExpenseLimitUsecase) Create(ctx context.Context, limit *entity.ExpenseLimit) error {
	return uc.expenseLimitRepo.Upsert(ctx, limit)
}

func (uc *ExpenseLimitUsecase) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, amount float64) error {
	limit, err := uc.expenseLimitRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if limit.UserID != userID {
		return domain.ErrForbidden
	}
	limit.Amount = amount
	return uc.expenseLimitRepo.Update(ctx, limit)
}

func (uc *ExpenseLimitUsecase) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	limit, err := uc.expenseLimitRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if limit.UserID != userID {
		return domain.ErrForbidden
	}
	return uc.expenseLimitRepo.Delete(ctx, id)
}
