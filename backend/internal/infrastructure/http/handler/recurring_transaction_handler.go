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

type RecurringTransactionHandler struct {
	uc *usecase.RecurringTransactionUsecase
}

func NewRecurringTransactionHandler(uc *usecase.RecurringTransactionUsecase) *RecurringTransactionHandler {
	return &RecurringTransactionHandler{uc: uc}
}

type recurringTransactionRequest struct {
	Type           string  `json:"type" binding:"required,oneof=income expense"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Description    string  `json:"description"`
	CategoryID     string  `json:"category_id" binding:"required,uuid"`
	Frequency      string  `json:"frequency" binding:"required,oneof=weekly biweekly monthly yearly"`
	StartDate      string  `json:"start_date" binding:"required"`
	EndDate        *string `json:"end_date"`
	MaxOccurrences *int    `json:"max_occurrences"`
	DayOfMonth     *int    `json:"day_of_month"`
}

type deleteRecurringRequest struct {
	Mode string `json:"mode" binding:"required,oneof=all future_and_current future_only"`
}

func (h *RecurringTransactionHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	filter := entity.RecurringTransactionFilter{
		Type: c.Query("type"),
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filter.IsActive = &isActive
	}

	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PerPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))

	result, err := h.uc.List(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *RecurringTransactionHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req recurringTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	catID, _ := uuid.Parse(req.CategoryID)
	rt := &entity.RecurringTransaction{
		UserID:         userID,
		CategoryID:     catID,
		Type:           req.Type,
		Amount:         req.Amount,
		Description:    req.Description,
		Frequency:      req.Frequency,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		MaxOccurrences: req.MaxOccurrences,
		DayOfMonth:     req.DayOfMonth,
	}

	if err := h.uc.Create(c.Request.Context(), rt); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rt)
}

func (h *RecurringTransactionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req deleteRecurringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id, entity.DeleteMode(req.Mode)); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *RecurringTransactionHandler) Pause(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Pause(c.Request.Context(), id); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "paused"})
}

func (h *RecurringTransactionHandler) Resume(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Resume(c.Request.Context(), id); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resumed"})
}
