package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/google/uuid"
)

type TransactionUsecase struct {
	transactionRepo repository.TransactionRepository
}

func NewTransactionUsecase(repo repository.TransactionRepository) *TransactionUsecase {
	return &TransactionUsecase{transactionRepo: repo}
}

func (uc *TransactionUsecase) List(ctx context.Context, filter entity.TransactionFilter) (*entity.PaginatedTransactions, error) {
	return uc.transactionRepo.FindAll(ctx, filter)
}

func (uc *TransactionUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	return uc.transactionRepo.FindByID(ctx, id)
}

func (uc *TransactionUsecase) Create(ctx context.Context, tx *entity.Transaction) error {
	return uc.transactionRepo.Create(ctx, tx)
}

func (uc *TransactionUsecase) Update(ctx context.Context, tx *entity.Transaction) error {
	if _, err := uc.transactionRepo.FindByID(ctx, tx.ID); err != nil {
		return err
	}
	return uc.transactionRepo.Update(ctx, tx)
}

func (uc *TransactionUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.transactionRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.transactionRepo.Delete(ctx, id)
}
