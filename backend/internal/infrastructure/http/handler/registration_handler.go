package handler

import (
	"errors"
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	uc *usecase.RegistrationUsecase
}

func NewRegistrationHandler(uc *usecase.RegistrationUsecase) *RegistrationHandler {
	return &RegistrationHandler{uc: uc}
}

type registerRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	TenantName string `json:"tenant_name" binding:"required,min=2"`
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *RegistrationHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.uc.Register(c.Request.Context(), usecase.RegisterInput{
		Name:       req.Name,
		Email:      req.Email,
		Password:   req.Password,
		TenantName: req.TenantName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{"error": "email já cadastrado"})
			return
		}
		if errors.Is(err, domain.ErrDuplicateTenant) {
			c.JSON(http.StatusConflict, gin.H{"error": "nome de dashboard já em uso"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar conta"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Conta criada! Verifique seu email para ativar."})
}

func (h *RegistrationHandler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.uc.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token inválido"})
			return
		}
		if errors.Is(err, domain.ErrInviteExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token expirado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao verificar email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verificado com sucesso!"})
}
