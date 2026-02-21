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
	userRepo       repository.UserRepository
	globalUserRepo repository.GlobalUserRepository
	membershipRepo repository.MembershipRepository
	jwtSecret      string
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	globalUserRepo repository.GlobalUserRepository,
	membershipRepo repository.MembershipRepository,
	jwtSecret string,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:       userRepo,
		globalUserRepo: globalUserRepo,
		membershipRepo: membershipRepo,
		jwtSecret:      jwtSecret,
	}
}

type LoginResult struct {
	// Single tenant: token + user + tenant_id returned directly
	Token    string    `json:"token,omitempty"`
	User     *entity.User `json:"user,omitempty"`
	TenantID *uuid.UUID   `json:"tenant_id,omitempty"`

	// Multi-tenant: selector_token + tenants list
	SelectorToken string                    `json:"selector_token,omitempty"`
	Tenants       []entity.TenantMembership `json:"tenants,omitempty"`
}

// AuthenticateGlobal validates credentials against global_users and returns either
// a full JWT (single tenant) or a selector token (multiple tenants).
func (uc *AuthUsecase) AuthenticateGlobal(ctx context.Context, email, password string) (*LoginResult, error) {
	globalUser, err := uc.globalUserRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(globalUser.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !globalUser.EmailVerified {
		return nil, domain.ErrEmailNotVerified
	}

	memberships, err := uc.membershipRepo.FindByGlobalUser(ctx, globalUser.ID)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return nil, domain.ErrNoMemberships
	}

	// Single tenant: auto-select
	if len(memberships) == 1 {
		token, schemaUserID, role, tenantID, err := uc.selectTenantInternal(ctx, globalUser.ID, memberships[0].TenantID)
		if err != nil {
			return nil, err
		}
		user := &entity.User{
			ID:   schemaUserID,
			Role: role,
		}
		return &LoginResult{Token: token, User: user, TenantID: &tenantID}, nil
	}

	// Multiple tenants: return selector token
	selectorToken, err := uc.generateSelectorToken(globalUser.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{SelectorToken: selectorToken, Tenants: memberships}, nil
}

// SelectTenant validates a selector token and returns a full JWT for the chosen tenant.
func (uc *AuthUsecase) SelectTenant(ctx context.Context, selectorToken string, tenantID uuid.UUID) (string, uuid.UUID, string, error) {
	claims, err := uc.parseSelectorToken(selectorToken)
	if err != nil {
		return "", uuid.Nil, "", domain.ErrInvalidCredentials
	}

	globalUserIDStr, _ := claims["global_user_id"].(string)
	globalUserID, err := uuid.Parse(globalUserIDStr)
	if err != nil {
		return "", uuid.Nil, "", domain.ErrInvalidCredentials
	}

	purpose, _ := claims["purpose"].(string)
	if purpose != "select" {
		return "", uuid.Nil, "", domain.ErrInvalidCredentials
	}

	token, schemaUserID, role, _, err := uc.selectTenantInternal(ctx, globalUserID, tenantID)
	if err != nil {
		return "", uuid.Nil, "", err
	}

	return token, schemaUserID, role, nil
}

func (uc *AuthUsecase) selectTenantInternal(ctx context.Context, globalUserID, tenantID uuid.UUID) (string, uuid.UUID, string, uuid.UUID, error) {
	membership, err := uc.membershipRepo.FindByGlobalUserAndTenant(ctx, globalUserID, tenantID)
	if err != nil {
		return "", uuid.Nil, "", uuid.Nil, domain.ErrForbidden
	}

	token, err := uc.generateToken(membership.SchemaUserID, tenantID, globalUserID, membership.Role)
	if err != nil {
		return "", uuid.Nil, "", uuid.Nil, err
	}

	return token, membership.SchemaUserID, membership.Role, tenantID, nil
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

func (uc *AuthUsecase) generateToken(schemaUserID, tenantID, globalUserID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":            schemaUserID.String(),
		"tenant_id":      tenantID.String(),
		"global_user_id": globalUserID.String(),
		"role":           role,
		"exp":            time.Now().Add(72 * time.Hour).Unix(),
		"iat":            time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUsecase) generateSelectorToken(globalUserID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"global_user_id": globalUserID.String(),
		"purpose":        "select",
		"exp":            time.Now().Add(5 * time.Minute).Unix(),
		"iat":            time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUsecase) parseSelectorToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(uc.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidCredentials
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}
	return claims, nil
}
