package handler

import (
	"errors"
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/dcunha/finance/backend/internal/tenant"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	uc          *usecase.AuthUsecase
	pool        *pgxpool.Pool
	tenantCache *database.TenantCache
}

func NewAuthHandler(uc *usecase.AuthUsecase, pool *pgxpool.Pool, tenantCache *database.TenantCache) *AuthHandler {
	return &AuthHandler{uc: uc, pool: pool, tenantCache: tenantCache}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type selectTenantRequest struct {
	SelectorToken string `json:"selector_token" binding:"required"`
	TenantID      string `json:"tenant_id" binding:"required"`
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

	result, err := h.uc.AuthenticateGlobal(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email ou senha inválidos"})
			return
		}
		if errors.Is(err, domain.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"error": "verifique seu email antes de fazer login"})
			return
		}
		if errors.Is(err, domain.ErrNoMemberships) {
			c.JSON(http.StatusForbidden, gin.H{"error": "conta sem acesso a nenhum dashboard"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Single tenant: enrich user with full schema data
	if result.Token != "" && result.User != nil && result.TenantID != nil {
		enriched := h.enrichUser(c, result.User.ID, *result.TenantID)
		if enriched != nil {
			result.User = enriched
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) SelectTenant(c *gin.Context) {
	var req selectTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	token, schemaUserID, role, err := h.uc.SelectTenant(c.Request.Context(), req.SelectorToken, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "sessão expirada, faça login novamente"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Enrich user with schema data
	user := h.enrichUser(c, schemaUserID, tenantID)
	if user == nil {
		user = &entity.User{ID: schemaUserID, Role: role}
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// enrichUser acquires a schema connection and fetches the full per-schema user.
func (h *AuthHandler) enrichUser(c *gin.Context, schemaUserID, tenantID uuid.UUID) *entity.User {
	t, ok := h.tenantCache.GetByID(tenantID)
	if !ok {
		return nil
	}

	ctx := tenant.ContextWithSchema(c.Request.Context(), t.SchemaName)
	conn, release, err := database.AcquireWithSchema(ctx, h.pool)
	if err != nil {
		return nil
	}
	defer release()
	ctx = database.ContextWithConn(ctx, conn)

	user, err := h.uc.GetProfile(ctx, schemaUserID)
	if err != nil {
		return nil
	}
	return user
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
