package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExpenseLimitHandler struct {
	uc *usecase.ExpenseLimitUsecase
}

func NewExpenseLimitHandler(uc *usecase.ExpenseLimitUsecase) *ExpenseLimitHandler {
	return &ExpenseLimitHandler{uc: uc}
}

type expenseLimitRequest struct {
	CategoryID *string `json:"category_id"`
	Month      int     `json:"month" binding:"required,min=1,max=12"`
	Year       int     `json:"year" binding:"required,min=2000"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

type updateLimitRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type copyLimitsRequest struct {
	FromMonth int `json:"from_month" binding:"required,min=1,max=12"`
	FromYear  int `json:"from_year" binding:"required,min=2000"`
	ToMonth   int `json:"to_month" binding:"required,min=1,max=12"`
	ToYear    int `json:"to_year" binding:"required,min=2000"`
}

func (h *ExpenseLimitHandler) List(c *gin.Context) {
	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	limits, err := h.uc.List(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, limits)
}

func (h *ExpenseLimitHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req expenseLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := &entity.ExpenseLimit{
		UserID: userID,
		Month:  req.Month,
		Year:   req.Year,
		Amount: req.Amount,
	}

	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		limit.CategoryID = &catID
	}

	if err := h.uc.Create(c.Request.Context(), limit); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, limit)
}

func (h *ExpenseLimitHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.Update(c.Request.Context(), id, req.Amount); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ExpenseLimitHandler) Copy(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req copyLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	copied, err := h.uc.CopyLimits(c.Request.Context(), req.FromMonth, req.FromYear, req.ToMonth, req.ToYear, userID)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"copied": copied})
}

func (h *ExpenseLimitHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
