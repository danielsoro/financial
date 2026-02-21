package handler

import (
	"errors"
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

type InviteHandler struct {
	uc *usecase.InviteUsecase
}

func NewInviteHandler(uc *usecase.InviteUsecase) *InviteHandler {
	return &InviteHandler{uc: uc}
}

type createInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin user"`
}

type acceptInviteRequest struct {
	Token    string `json:"token" binding:"required"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (h *InviteHandler) CreateInvite(c *gin.Context) {
	var req createInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)
	globalUserID := middleware.GetGlobalUserID(c)

	err := h.uc.CreateInvite(c.Request.Context(), tenantID, globalUserID, req.Email, req.Role)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyMember) {
			c.JSON(http.StatusConflict, gin.H{"error": "usuário já é membro deste dashboard"})
			return
		}
		if errors.Is(err, domain.ErrInvalidRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "papel inválido"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao enviar convite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Convite enviado!"})
}

func (h *InviteHandler) GetInviteInfo(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	info, err := h.uc.GetInviteInfo(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "convite não encontrado"})
			return
		}
		if errors.Is(err, domain.ErrInviteExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "convite expirado"})
			return
		}
		if errors.Is(err, domain.ErrInviteAlreadyUsed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "convite já utilizado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *InviteHandler) AcceptInvite(c *gin.Context) {
	var req acceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.uc.AcceptInvite(c.Request.Context(), usecase.AcceptInviteInput{
		Token:    req.Token,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "convite não encontrado"})
			return
		}
		if errors.Is(err, domain.ErrInviteExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "convite expirado"})
			return
		}
		if errors.Is(err, domain.ErrInviteAlreadyUsed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "convite já utilizado"})
			return
		}
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nome e senha são obrigatórios para novos usuários"})
			return
		}
		if errors.Is(err, domain.ErrAlreadyMember) {
			c.JSON(http.StatusConflict, gin.H{"error": "você já é membro deste dashboard"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao aceitar convite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Convite aceito! Faça login para acessar."})
}
