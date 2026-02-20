package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	uc *usecase.DashboardUsecase
}

func NewDashboardHandler(uc *usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	month, year := getMonthYear(c)

	summary, err := h.uc.GetSummary(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *DashboardHandler) ByCategory(c *gin.Context) {
	month, year := getMonthYear(c)
	txType := c.DefaultQuery("type", "expense")

	data, err := h.uc.GetByCategory(c.Request.Context(), month, year, txType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) LimitsProgress(c *gin.Context) {
	month, year := getMonthYear(c)

	progress, err := h.uc.GetLimitsProgress(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

func getMonthYear(c *gin.Context) (int, int) {
	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	return month, year
}
