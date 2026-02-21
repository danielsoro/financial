package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type MembershipRepository interface {
	Create(ctx context.Context, membership *entity.Membership) error
	FindByGlobalUser(ctx context.Context, globalUserID uuid.UUID) ([]entity.TenantMembership, error)
	FindByGlobalUserAndTenant(ctx context.Context, globalUserID, tenantID uuid.UUID) (*entity.Membership, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
