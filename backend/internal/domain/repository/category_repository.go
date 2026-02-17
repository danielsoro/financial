package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, cat *entity.Category) error
	Update(ctx context.Context, cat *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	FindAllForTenant(ctx context.Context, tenantID uuid.UUID, catType string) ([]entity.Category, error)
	IsInUse(ctx context.Context, id uuid.UUID) (bool, error)
	IsSubtreeInUse(ctx context.Context, id uuid.UUID) (bool, error)
}
