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

func (uc *ExpenseLimitUsecase) List(ctx context.Context, month, year int) ([]entity.ExpenseLimit, error) {
	return uc.expenseLimitRepo.FindAll(ctx, month, year)
}

func (uc *ExpenseLimitUsecase) Create(ctx context.Context, limit *entity.ExpenseLimit) error {
	return uc.expenseLimitRepo.Upsert(ctx, limit)
}

func (uc *ExpenseLimitUsecase) Update(ctx context.Context, id uuid.UUID, amount float64) error {
	limit, err := uc.expenseLimitRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	limit.Amount = amount
	return uc.expenseLimitRepo.Update(ctx, limit)
}

func (uc *ExpenseLimitUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.expenseLimitRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.expenseLimitRepo.Delete(ctx, id)
}

func (uc *ExpenseLimitUsecase) CopyLimits(ctx context.Context, fromMonth, fromYear, toMonth, toYear int, userID uuid.UUID) (int, error) {
	if fromMonth == toMonth && fromYear == toYear {
		return 0, domain.ErrSameMonth
	}
	limits, err := uc.expenseLimitRepo.FindAll(ctx, fromMonth, fromYear)
	if err != nil {
		return 0, err
	}
	if len(limits) == 0 {
		return 0, domain.ErrNotFound
	}
	for _, limit := range limits {
		newLimit := &entity.ExpenseLimit{
			UserID:     userID,
			CategoryID: limit.CategoryID,
			Month:      toMonth,
			Year:       toYear,
			Amount:     limit.Amount,
		}
		if err := uc.expenseLimitRepo.Upsert(ctx, newLimit); err != nil {
			return 0, err
		}
	}
	return len(limits), nil
}
