package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain"
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

func (uc *TransactionUsecase) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Transaction, error) {
	tx, err := uc.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tx.TenantID != tenantID {
		return nil, domain.ErrForbidden
	}
	return tx, nil
}

func (uc *TransactionUsecase) Create(ctx context.Context, tx *entity.Transaction) error {
	return uc.transactionRepo.Create(ctx, tx)
}

func (uc *TransactionUsecase) Update(ctx context.Context, tenantID uuid.UUID, tx *entity.Transaction) error {
	existing, err := uc.transactionRepo.FindByID(ctx, tx.ID)
	if err != nil {
		return err
	}
	if existing.TenantID != tenantID {
		return domain.ErrForbidden
	}
	return uc.transactionRepo.Update(ctx, tx)
}

func (uc *TransactionUsecase) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	tx, err := uc.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if tx.TenantID != tenantID {
		return domain.ErrForbidden
	}
	return uc.transactionRepo.Delete(ctx, id)
}
