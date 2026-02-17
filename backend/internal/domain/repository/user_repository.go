package repository

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	FindAllByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AdminUser, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
