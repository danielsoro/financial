package usecase

import (
	"context"
	"time"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	jwtSecret  string
}

func NewAuthUsecase(userRepo repository.UserRepository, tenantRepo repository.TenantRepository, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, tenantRepo: tenantRepo, jwtSecret: jwtSecret}
}

// ResolveTenant looks up the tenant by subdomain (public schema, no schema-scoped connection needed).
func (uc *AuthUsecase) ResolveTenant(ctx context.Context, subdomain string) (*entity.Tenant, error) {
	if subdomain == "" {
		subdomain = "root"
	}
	t, err := uc.tenantRepo.FindByDomain(ctx, subdomain)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if !t.IsActive {
		return nil, domain.ErrForbidden
	}
	return t, nil
}

// Authenticate validates credentials and returns a JWT token.
// Expects a schema-scoped connection to be already in the context.
func (uc *AuthUsecase) Authenticate(ctx context.Context, email, password string, tenantID uuid.UUID) (string, *entity.User, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return "", nil, domain.ErrInvalidCredentials
		}
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := uc.generateToken(user, tenantID)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (uc *AuthUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *AuthUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, name, email string) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.Email = email
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *AuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	return uc.userRepo.Update(ctx, user)
}

func (uc *AuthUsecase) generateToken(user *entity.User, tenantID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"tenant_id": tenantID.String(),
		"role":      user.Role,
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
