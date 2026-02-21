package main

import (
	"context"
	"log"

	"github.com/dcunha/finance/backend/internal/config"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/infrastructure/email"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/handler"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/router"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run public migrations (creates tenants, global_users, memberships, invites tables)
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize all existing tenant schemas (run tenant migrations)
	sm := database.NewSchemaManager(pool)
	if err := sm.InitAllTenants(ctx, cfg.DatabaseURL, "tenant_migrations"); err != nil {
		log.Fatalf("Failed to initialize tenant schemas: %v", err)
	}

	// Load tenant cache
	tenantCache := database.NewTenantCache()
	if err := tenantCache.Load(ctx, pool); err != nil {
		log.Fatalf("Failed to load tenant cache: %v", err)
	}

	// Email sender
	emailSender := email.NewSender(cfg.SendGridAPIKey, cfg.EmailFrom)

	// Repositories
	tenantRepo := database.NewTenantRepo(pool)
	userRepo := database.NewUserRepo()
	categoryRepo := database.NewCategoryRepo()
	transactionRepo := database.NewTransactionRepo()
	expenseLimitRepo := database.NewExpenseLimitRepo()
	recurringRepo := database.NewRecurringTransactionRepo()
	globalUserRepo := database.NewGlobalUserRepo(pool)
	membershipRepo := database.NewMembershipRepo(pool)
	inviteRepo := database.NewInviteRepo(pool)

	// Usecases
	healthUc := usecase.NewHealthUsecase(pool)
	authUC := usecase.NewAuthUsecase(userRepo, globalUserRepo, membershipRepo, cfg.JWTSecret)
	adminUC := usecase.NewAdminUsecase(userRepo)
	categoryUC := usecase.NewCategoryUsecase(categoryRepo)
	transactionUC := usecase.NewTransactionUsecase(transactionRepo)
	expenseLimitUC := usecase.NewExpenseLimitUsecase(expenseLimitRepo)
	dashboardUC := usecase.NewDashboardUsecase(transactionRepo, expenseLimitRepo)
	recurringUC := usecase.NewRecurringTransactionUsecase(recurringRepo, transactionRepo)
	registrationUC := usecase.NewRegistrationUsecase(
		globalUserRepo, membershipRepo, tenantRepo, userRepo,
		sm, tenantCache, pool, emailSender,
		cfg.AppURL, cfg.DatabaseURL, "tenant_migrations",
	)
	inviteUC := usecase.NewInviteUsecase(
		inviteRepo, globalUserRepo, membershipRepo, tenantRepo,
		registrationUC, tenantCache, emailSender, cfg.AppURL,
	)

	// Handlers
	handlers := router.Handlers{
		Health:       handler.NewHealthHandler(healthUc),
		Auth:         handler.NewAuthHandler(authUC, pool, tenantCache),
		Registration: handler.NewRegistrationHandler(registrationUC),
		Invite:       handler.NewInviteHandler(inviteUC),
		Admin:        handler.NewAdminHandler(adminUC),
		Category:     handler.NewCategoryHandler(categoryUC),
		Transaction:  handler.NewTransactionHandler(transactionUC),
		ExpenseLimit: handler.NewExpenseLimitHandler(expenseLimitUC),
		Dashboard:    handler.NewDashboardHandler(dashboardUC),
		Recurring:    handler.NewRecurringTransactionHandler(recurringUC),
	}

	// Router
	r := gin.Default()
	r.TrustedPlatform = gin.PlatformCloudflare
	router.Setup(r, cfg.JWTSecret, cfg.StaticDir, cfg.AllowedOrigin, pool, tenantCache, handlers)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
