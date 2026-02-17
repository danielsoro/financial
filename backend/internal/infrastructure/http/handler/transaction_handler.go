package handler

import (
	"net/http"
	"strconv"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	uc *usecase.TransactionUsecase
}

func NewTransactionHandler(uc *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

type transactionRequest struct {
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
	CategoryID  string  `json:"category_id" binding:"required,uuid"`
}

func (h *TransactionHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	filter := entity.TransactionFilter{
		TenantID:  tenantID,
		Type:      c.Query("type"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	if catID := c.Query("category_id"); catID != "" {
		id, err := uuid.Parse(catID)
		if err == nil {
			filter.CategoryID = &id
		}
	}

	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PerPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))

	result, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tx, err := h.uc.GetByID(c.Request.Context(), tenantID, id)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	var req transactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	catID, _ := uuid.Parse(req.CategoryID)
	tx := &entity.Transaction{
		TenantID:    tenantID,
		UserID:      userID,
		CategoryID:  catID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        req.Date,
	}

	if err := h.uc.Create(c.Request.Context(), tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

func (h *TransactionHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req transactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	catID, _ := uuid.Parse(req.CategoryID)
	tx := &entity.Transaction{
		ID:          id,
		CategoryID:  catID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        req.Date,
	}

	if err := h.uc.Update(c.Request.Context(), tenantID, tx); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), tenantID, id); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
