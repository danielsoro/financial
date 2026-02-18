package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminUsecase struct {
	userRepo repository.UserRepository
}

func NewAdminUsecase(userRepo repository.UserRepository) *AdminUsecase {
	return &AdminUsecase{userRepo: userRepo}
}

func (uc *AdminUsecase) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]entity.AdminUser, error) {
	return uc.userRepo.FindAllByTenant(ctx, tenantID)
}

func (uc *AdminUsecase) CreateUser(ctx context.Context, tenantID uuid.UUID, name, email, password, role string) (*entity.AdminUser, error) {
	if role != "admin" && role != "user" {
		return nil, domain.ErrInvalidRole
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &entity.User{
		TenantID:     tenantID,
		Name:         name,
		Email:        email,
		Role:         role,
		PasswordHash: string(hash),
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return &entity.AdminUser{
		ID:        user.ID,
		TenantID:  user.TenantID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (uc *AdminUsecase) UpdateUser(ctx context.Context, id uuid.UUID, name, email, role string) (*entity.AdminUser, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != "admin" && role != "user" {
		return nil, domain.ErrInvalidRole
	}
	user.Name = name
	user.Email = email
	user.Role = role
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return &entity.AdminUser{
		ID:        user.ID,
		TenantID:  user.TenantID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (uc *AdminUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.userRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.userRepo.DeleteUser(ctx, id)
}

func (uc *AdminUsecase) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	return uc.userRepo.Update(ctx, user)
}
