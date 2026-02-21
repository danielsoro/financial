package usecase

import (
	"context"
	"time"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/infrastructure/email"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type InviteUsecase struct {
	inviteRepo     repository.InviteRepository
	globalUserRepo repository.GlobalUserRepository
	membershipRepo repository.MembershipRepository
	tenantRepo     repository.TenantRepository
	regUC          *RegistrationUsecase
	tenantCache    *database.TenantCache
	emailSender    email.Sender
	appURL         string
}

func NewInviteUsecase(
	inviteRepo repository.InviteRepository,
	globalUserRepo repository.GlobalUserRepository,
	membershipRepo repository.MembershipRepository,
	tenantRepo repository.TenantRepository,
	regUC *RegistrationUsecase,
	tenantCache *database.TenantCache,
	emailSender email.Sender,
	appURL string,
) *InviteUsecase {
	return &InviteUsecase{
		inviteRepo:     inviteRepo,
		globalUserRepo: globalUserRepo,
		membershipRepo: membershipRepo,
		tenantRepo:     tenantRepo,
		regUC:          regUC,
		tenantCache:    tenantCache,
		emailSender:    emailSender,
		appURL:         appURL,
	}
}

func (uc *InviteUsecase) CreateInvite(ctx context.Context, tenantID, invitedByGlobalUserID uuid.UUID, emailAddr, role string) error {
	if role != "admin" && role != "user" {
		return domain.ErrInvalidRole
	}

	// Check if already a member
	existing, _ := uc.globalUserRepo.FindByEmail(ctx, emailAddr)
	if existing != nil {
		_, err := uc.membershipRepo.FindByGlobalUserAndTenant(ctx, existing.ID, tenantID)
		if err == nil {
			return domain.ErrAlreadyMember
		}
	}

	token, err := generateRandomToken()
	if err != nil {
		return err
	}

	invite := &entity.Invite{
		TenantID:  tenantID,
		Email:     emailAddr,
		Role:      role,
		Token:     token,
		InvitedBy: invitedByGlobalUserID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := uc.inviteRepo.Create(ctx, invite); err != nil {
		return err
	}

	// Send invite email
	t, _ := uc.tenantRepo.FindByID(ctx, tenantID)
	inviter, _ := uc.globalUserRepo.FindByID(ctx, invitedByGlobalUserID)
	tenantName := "Dashboard"
	inviterName := "Um administrador"
	if t != nil {
		tenantName = t.Name
	}
	if inviter != nil {
		inviterName = inviter.Name
	}

	subject, body := email.InviteEmail(uc.appURL, token, tenantName, inviterName)
	uc.emailSender.Send(emailAddr, subject, body)

	return nil
}

func (uc *InviteUsecase) GetInviteInfo(ctx context.Context, token string) (*entity.InviteInfo, error) {
	invite, err := uc.inviteRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if invite.AcceptedAt != nil {
		return nil, domain.ErrInviteAlreadyUsed
	}

	if invite.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrInviteExpired
	}

	t, err := uc.tenantRepo.FindByID(ctx, invite.TenantID)
	if err != nil {
		return nil, err
	}

	existing, _ := uc.globalUserRepo.FindByEmail(ctx, invite.Email)
	userExists := existing != nil

	return &entity.InviteInfo{
		TenantName: t.Name,
		Email:      invite.Email,
		Role:       invite.Role,
		UserExists: userExists,
	}, nil
}

type AcceptInviteInput struct {
	Token    string
	Name     string
	Password string
}

func (uc *InviteUsecase) AcceptInvite(ctx context.Context, input AcceptInviteInput) error {
	invite, err := uc.inviteRepo.FindByToken(ctx, input.Token)
	if err != nil {
		return domain.ErrNotFound
	}

	if invite.AcceptedAt != nil {
		return domain.ErrInviteAlreadyUsed
	}

	if invite.ExpiresAt.Before(time.Now()) {
		return domain.ErrInviteExpired
	}

	t, err := uc.tenantRepo.FindByID(ctx, invite.TenantID)
	if err != nil {
		return err
	}

	// Check if user already exists
	globalUser, _ := uc.globalUserRepo.FindByEmail(ctx, invite.Email)

	if globalUser == nil {
		// New user: create global account (verified since they received the invite email)
		if input.Name == "" || input.Password == "" {
			return domain.ErrInvalidCredentials
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		globalUser = &entity.GlobalUser{
			Name:            input.Name,
			Email:           invite.Email,
			PasswordHash:    string(hash),
			EmailVerified:   true,
			MaxOwnedTenants: 1,
		}
		if err := uc.globalUserRepo.Create(ctx, globalUser); err != nil {
			return err
		}
	}

	// Create per-schema user
	passwordHash := globalUser.PasswordHash
	schemaUser, err := uc.regUC.CreateSchemaUser(ctx, t.SchemaName, globalUser.Name, globalUser.Email, passwordHash, invite.Role, globalUser.ID)
	if err != nil {
		return err
	}

	// Create membership
	membership := &entity.Membership{
		GlobalUserID: globalUser.ID,
		TenantID:     invite.TenantID,
		SchemaUserID: schemaUser.ID,
		Role:         invite.Role,
	}
	if err := uc.membershipRepo.Create(ctx, membership); err != nil {
		return err
	}

	// Mark invite as accepted
	return uc.inviteRepo.MarkAccepted(ctx, invite.ID)
}
