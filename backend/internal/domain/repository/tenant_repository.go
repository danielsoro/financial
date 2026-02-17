package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*entity.Tenant, error)
	FindAll(ctx context.Context) ([]entity.Tenant, error)
}
