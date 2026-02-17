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

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Repositories
	userRepo := database.NewUserRepo(pool)
	categoryRepo := database.NewCategoryRepo(pool)
	transactionRepo := database.NewTransactionRepo(pool)
	expenseLimitRepo := database.NewExpenseLimitRepo(pool)

	// Usecases
	authUC := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)
	categoryUC := usecase.NewCategoryUsecase(categoryRepo)
	transactionUC := usecase.NewTransactionUsecase(transactionRepo)
	expenseLimitUC := usecase.NewExpenseLimitUsecase(expenseLimitRepo)
	dashboardUC := usecase.NewDashboardUsecase(transactionRepo, expenseLimitRepo)

	// Handlers
	handlers := router.Handlers{
		Auth:         handler.NewAuthHandler(authUC),
		Category:     handler.NewCategoryHandler(categoryUC),
		Transaction:  handler.NewTransactionHandler(transactionUC),
		ExpenseLimit: handler.NewExpenseLimitHandler(expenseLimitUC),
		Dashboard:    handler.NewDashboardHandler(dashboardUC),
	}

	// Router
	r := gin.Default()
	router.Setup(r, cfg.JWTSecret, handlers)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
