package handler

import (
	"errors"
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/dcunha/finance/backend/internal/tenant"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	uc   *usecase.AuthUsecase
	pool *pgxpool.Pool
}

func NewAuthHandler(uc *usecase.AuthUsecase, pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{uc: uc, pool: pool}
}

type loginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	Subdomain string `json:"subdomain"`
}

type updateProfileRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Resolve tenant (public schema — no schema-scoped connection needed)
	t, err := h.uc.ResolveTenant(c.Request.Context(), req.Subdomain)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 2. Set up schema-scoped connection for authentication queries
	ctx := tenant.ContextWithSchema(c.Request.Context(), t.SchemaName)
	conn, release, err := database.AcquireWithSchema(ctx, h.pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer release()
	ctx = database.ContextWithConn(ctx, conn)

	// 3. Authenticate (UserRepo uses connection from context)
	token, user, err := h.uc.Authenticate(ctx, req.Email, req.Password, t.ID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.uc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	user, err := h.uc.UpdateProfile(c.Request.Context(), userID, req.Name, req.Email)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.uc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
