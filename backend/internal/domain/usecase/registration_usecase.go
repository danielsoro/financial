package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/infrastructure/email"
	"github.com/dcunha/finance/backend/internal/tenant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

type RegistrationUsecase struct {
	globalUserRepo repository.GlobalUserRepository
	membershipRepo repository.MembershipRepository
	tenantRepo     repository.TenantRepository
	userRepo       repository.UserRepository
	schemaManager  *database.SchemaManager
	tenantCache    *database.TenantCache
	pool           *pgxpool.Pool
	emailSender    email.Sender
	appURL         string
	databaseURL    string
	migrationsDir  string
}

func NewRegistrationUsecase(
	globalUserRepo repository.GlobalUserRepository,
	membershipRepo repository.MembershipRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	schemaManager *database.SchemaManager,
	tenantCache *database.TenantCache,
	pool *pgxpool.Pool,
	emailSender email.Sender,
	appURL, databaseURL, migrationsDir string,
) *RegistrationUsecase {
	return &RegistrationUsecase{
		globalUserRepo: globalUserRepo,
		membershipRepo: membershipRepo,
		tenantRepo:     tenantRepo,
		userRepo:       userRepo,
		schemaManager:  schemaManager,
		tenantCache:    tenantCache,
		pool:           pool,
		emailSender:    emailSender,
		appURL:         appURL,
		databaseURL:    databaseURL,
		migrationsDir:  migrationsDir,
	}
}

type RegisterInput struct {
	Name       string
	Email      string
	Password   string
	TenantName string
}

func (uc *RegistrationUsecase) Register(ctx context.Context, input RegisterInput) error {
	// Check email uniqueness
	existing, _ := uc.globalUserRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return domain.ErrDuplicateEmail
	}

	// Generate slug for tenant
	slug := generateSlug(input.TenantName)
	schemaName := "tenant_" + slug

	// Ensure unique schema_name
	schemaName, slug, err := uc.ensureUniqueSchema(ctx, schemaName, slug)
	if err != nil {
		return err
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Generate email token
	emailToken, err := generateRandomToken()
	if err != nil {
		return err
	}
	tokenExpiry := time.Now().Add(24 * time.Hour)

	// Create global user
	globalUser := &entity.GlobalUser{
		Name:                input.Name,
		Email:               input.Email,
		PasswordHash:        string(hash),
		EmailVerified:       false,
		EmailToken:          &emailToken,
		EmailTokenExpiresAt: &tokenExpiry,
		MaxOwnedTenants:     1,
	}
	if err := uc.globalUserRepo.Create(ctx, globalUser); err != nil {
		return err
	}

	// Create tenant
	t := &entity.Tenant{
		Name:       input.TenantName,
		Domain:     &slug,
		SchemaName: schemaName,
		IsActive:   true,
		OwnerID:    &globalUser.ID,
	}
	if err := uc.tenantRepo.Create(ctx, t); err != nil {
		return err
	}

	// Create schema + run tenant migrations
	if err := uc.schemaManager.InitTenantSchema(ctx, uc.databaseURL, uc.migrationsDir, schemaName); err != nil {
		return fmt.Errorf("initializing tenant schema: %w", err)
	}

	// Create per-schema user (needs schema connection)
	schemaCtx := tenant.ContextWithSchema(ctx, schemaName)
	conn, release, err := database.AcquireWithSchema(schemaCtx, uc.pool)
	if err != nil {
		return fmt.Errorf("acquiring schema connection: %w", err)
	}
	defer release()
	schemaCtx = database.ContextWithConn(schemaCtx, conn)

	schemaUser := &entity.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         "owner",
		GlobalUserID: &globalUser.ID,
	}
	if err := uc.userRepo.Create(schemaCtx, schemaUser); err != nil {
		return fmt.Errorf("creating schema user: %w", err)
	}

	// Create membership
	membership := &entity.Membership{
		GlobalUserID: globalUser.ID,
		TenantID:     t.ID,
		SchemaUserID: schemaUser.ID,
		Role:         "owner",
	}
	if err := uc.membershipRepo.Create(ctx, membership); err != nil {
		return fmt.Errorf("creating membership: %w", err)
	}

	// Add tenant to cache
	uc.tenantCache.Add(t)

	// Send verification email
	subject, body := email.VerificationEmail(uc.appURL, emailToken)
	if err := uc.emailSender.Send(input.Email, subject, body); err != nil {
		// Log but don't fail registration
		fmt.Printf("Warning: failed to send verification email to %s: %v\n", input.Email, err)
	}

	return nil
}

func (uc *RegistrationUsecase) VerifyEmail(ctx context.Context, token string) error {
	user, err := uc.globalUserRepo.FindByEmailToken(ctx, token)
	if err != nil {
		return domain.ErrNotFound
	}

	if user.EmailTokenExpiresAt != nil && user.EmailTokenExpiresAt.Before(time.Now()) {
		return domain.ErrInviteExpired
	}

	user.EmailVerified = true
	user.EmailToken = nil
	user.EmailTokenExpiresAt = nil
	return uc.globalUserRepo.Update(ctx, user)
}

func (uc *RegistrationUsecase) ensureUniqueSchema(ctx context.Context, schemaName, slug string) (string, string, error) {
	base := schemaName
	baseSlug := slug
	for i := 2; ; i++ {
		_, err := uc.tenantRepo.FindBySchemaName(ctx, schemaName)
		if err == domain.ErrTenantNotFound {
			return schemaName, slug, nil
		}
		if err != nil {
			return "", "", err
		}
		schemaName = fmt.Sprintf("%s_%d", base, i)
		slug = fmt.Sprintf("%s_%d", baseSlug, i)
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = slugRegex.ReplaceAllString(slug, "_")
	slug = strings.Trim(slug, "_")
	if slug == "" {
		slug = "default"
	}
	return slug
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSchemaUser creates a per-schema user. Used by registration and invite flows.
func (uc *RegistrationUsecase) CreateSchemaUser(ctx context.Context, schemaName string, name, emailAddr, passwordHash, role string, globalUserID uuid.UUID) (*entity.User, error) {
	schemaCtx := tenant.ContextWithSchema(ctx, schemaName)
	conn, release, err := database.AcquireWithSchema(schemaCtx, uc.pool)
	if err != nil {
		return nil, fmt.Errorf("acquiring schema connection: %w", err)
	}
	defer release()
	schemaCtx = database.ContextWithConn(schemaCtx, conn)

	schemaUser := &entity.User{
		Name:         name,
		Email:        emailAddr,
		PasswordHash: passwordHash,
		Role:         role,
		GlobalUserID: &globalUserID,
	}
	if err := uc.userRepo.Create(schemaCtx, schemaUser); err != nil {
		return nil, err
	}
	return schemaUser, nil
}
