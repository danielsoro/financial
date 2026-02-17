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
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthUsecase(repo repository.UserRepository, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{userRepo: repo, jwtSecret: jwtSecret}
}

func (uc *AuthUsecase) Register(ctx context.Context, name, email, password string) (string, *entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}
	user := &entity.User{Name: name, Email: email, PasswordHash: string(hash)}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return "", nil, err
	}
	token, err := uc.generateToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (string, *entity.User, error) {
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
	token, err := uc.generateToken(user)
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

func (uc *AuthUsecase) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(72 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
