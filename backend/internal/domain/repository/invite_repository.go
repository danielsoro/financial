package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *entity.Invite) error
	FindByToken(ctx context.Context, token string) (*entity.Invite, error)
	FindByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.Invite, error)
	MarkAccepted(ctx context.Context, id uuid.UUID) error
	FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Invite, error)
}
