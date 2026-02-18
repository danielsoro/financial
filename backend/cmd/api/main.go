package main

import (
	"context"
	"log"

	"github.com/dcunha/finance/backend/internal/config"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/database"
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

	// Run public migrations (creates tenants table)
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Ensure tenants from TENANTS env var exist in the tenants table
	if err := database.EnsureTenantsFromEnv(ctx, pool, cfg.Tenants); err != nil {
		log.Fatalf("Failed to ensure tenants: %v", err)
	}

	// Initialize all tenant schemas (create schemas, run tenant migrations, seed admin)
	sm := database.NewSchemaManager(pool)
	if err := sm.InitAllTenants(ctx, cfg.DatabaseURL, "tenant_migrations"); err != nil {
		log.Fatalf("Failed to initialize tenant schemas: %v", err)
	}

	// Load tenant cache (after all tenants are created)
	tenantCache := database.NewTenantCache()
	if err := tenantCache.Load(ctx, pool); err != nil {
		log.Fatalf("Failed to load tenant cache: %v", err)
	}

	// Repositories
	tenantRepo := database.NewTenantRepo(pool)
	userRepo := database.NewUserRepo(pool)
	categoryRepo := database.NewCategoryRepo(pool)
	transactionRepo := database.NewTransactionRepo(pool)
	expenseLimitRepo := database.NewExpenseLimitRepo(pool)

	// Usecases
	healthUc := usecase.NewHealthUsecase(pool)
	authUC := usecase.NewAuthUsecase(userRepo, tenantRepo, cfg.JWTSecret)
	adminUC := usecase.NewAdminUsecase(userRepo)
	categoryUC := usecase.NewCategoryUsecase(categoryRepo)
	transactionUC := usecase.NewTransactionUsecase(transactionRepo)
	expenseLimitUC := usecase.NewExpenseLimitUsecase(expenseLimitRepo)
	dashboardUC := usecase.NewDashboardUsecase(transactionRepo, expenseLimitRepo)

	// Handlers
	handlers := router.Handlers{
		Health:       handler.NewHealthHandler(healthUc),
		Auth:         handler.NewAuthHandler(authUC),
		Admin:        handler.NewAdminHandler(adminUC),
		Category:     handler.NewCategoryHandler(categoryUC),
		Transaction:  handler.NewTransactionHandler(transactionUC),
		ExpenseLimit: handler.NewExpenseLimitHandler(expenseLimitUC),
		Dashboard:    handler.NewDashboardHandler(dashboardUC),
	}

	// Router
	r := gin.Default()
	router.Setup(r, cfg.JWTSecret, cfg.StaticDir, cfg.AllowedOrigin, tenantCache, handlers)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
