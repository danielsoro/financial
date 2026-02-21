package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type GlobalUserRepository interface {
	Create(ctx context.Context, user *entity.GlobalUser) error
	FindByEmail(ctx context.Context, email string) (*entity.GlobalUser, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.GlobalUser, error)
	Update(ctx context.Context, user *entity.GlobalUser) error
	FindByEmailToken(ctx context.Context, token string) (*entity.GlobalUser, error)
	CountOwnedTenants(ctx context.Context, globalUserID uuid.UUID) (int, error)
}
