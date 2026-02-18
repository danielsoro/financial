package handler

import (
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	uc *usecase.HealthUsecase
}

func NewHealthHandler(uc *usecase.HealthUsecase) *HealthHandler {
	return &HealthHandler{uc}
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *HealthHandler) Health(c *gin.Context) {
	if err := h.uc.Check(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, healthResponse{Status: "NOK"})
		return
	}

	c.JSON(http.StatusOK, healthResponse{Status: "OK"})
}
