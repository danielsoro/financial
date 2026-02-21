package router

import (
	"net/http"
	"strings"

	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/handler"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	Health       *handler.HealthHandler
	Auth         *handler.AuthHandler
	Category     *handler.CategoryHandler
	Transaction  *handler.TransactionHandler
	ExpenseLimit *handler.ExpenseLimitHandler
	Dashboard    *handler.DashboardHandler
	Admin        *handler.AdminHandler
	Recurring    *handler.RecurringTransactionHandler
}

func Setup(r *gin.Engine, jwtSecret string, staticDir string, allowedOrigin string, pool *pgxpool.Pool, tenantCache *database.TenantCache, h Handlers) {
	r.Use(middleware.CORS(allowedOrigin))

	r.GET("/health", h.Health.Health)

	api := r.Group("/api/v1")

	// Auth (public)
	auth := api.Group("/auth")
	auth.POST("/login", h.Auth.Login)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtSecret, tenantCache))
	protected.Use(middleware.SchemaConn(pool))

	// Profile
	protected.GET("/profile", h.Auth.GetProfile)
	protected.PUT("/profile", h.Auth.UpdateProfile)
	protected.POST("/profile/change-password", h.Auth.ChangePassword)

	// Categories
	cats := protected.Group("/categories")
	cats.GET("", h.Category.List)
	cats.POST("", h.Category.Create)
	cats.PUT("/:id", h.Category.Update)
	cats.DELETE("/:id", h.Category.Delete)

	// Transactions
	txs := protected.Group("/transactions")
	txs.GET("", h.Transaction.List)
	txs.GET("/:id", h.Transaction.GetByID)
	txs.POST("", h.Transaction.Create)
	txs.PUT("/:id", h.Transaction.Update)
	txs.DELETE("/:id", h.Transaction.Delete)

	// Expense Limits
	limits := protected.Group("/expense-limits")
	limits.GET("", h.ExpenseLimit.List)
	limits.POST("", h.ExpenseLimit.Create)
	limits.POST("/copy", h.ExpenseLimit.Copy)
	limits.PUT("/:id", h.ExpenseLimit.Update)
	limits.DELETE("/:id", h.ExpenseLimit.Delete)

	// Recurring Transactions
	recurring := protected.Group("/recurring-transactions")
	recurring.GET("", h.Recurring.List)
	recurring.POST("", h.Recurring.Create)
	recurring.DELETE("/:id", h.Recurring.Delete)
	recurring.POST("/:id/pause", h.Recurring.Pause)
	recurring.POST("/:id/resume", h.Recurring.Resume)

	// Dashboard
	dash := protected.Group("/dashboard")
	dash.GET("/summary", h.Dashboard.Summary)
	dash.GET("/by-category", h.Dashboard.ByCategory)
	dash.GET("/limits-progress", h.Dashboard.LimitsProgress)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	admin.GET("/users", h.Admin.ListUsers)
	admin.POST("/users", h.Admin.CreateUser)
	admin.PUT("/users/:id", h.Admin.UpdateUser)
	admin.DELETE("/users/:id", h.Admin.DeleteUser)
	admin.POST("/users/:id/reset-password", h.Admin.ResetPassword)

	// Serve frontend static files (production)
	if staticDir != "" {
		r.StaticFS("/assets", http.Dir(staticDir+"/assets"))

		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(staticDir + "/index.html")
		})
	}
}
