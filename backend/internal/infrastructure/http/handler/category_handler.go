package handler

import (
	"errors"
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/usecase"
	"github.com/dcunha/finance/backend/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	uc *usecase.CategoryUsecase
}

func NewCategoryHandler(uc *usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

type categoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	Type     string  `json:"type" binding:"required,oneof=income expense both"`
	ParentID *string `json:"parent_id"`
}

func (h *CategoryHandler) List(c *gin.Context) {
	catType := c.Query("type")
	view := c.DefaultQuery("view", "flat")

	if view == "tree" {
		categories, err := h.uc.ListTree(c.Request.Context(), catType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, categories)
		return
	}

	categories, err := h.uc.List(c.Request.Context(), catType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsed, err := uuid.Parse(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_id"})
			return
		}
		parentID = &parsed
	}

	cat, err := h.uc.Create(c.Request.Context(), userID, req.Name, req.Type, parentID)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateCategory) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cat)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsed, err := uuid.Parse(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_id"})
			return
		}
		parentID = &parsed
	}

	cat, err := h.uc.Update(c.Request.Context(), id, req.Name, req.Type, parentID)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cat)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
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

func mapDomainError(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrCategoryInUse):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDuplicateCategory):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDuplicateEmail):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDuplicateLimit):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDuplicateDomain):
		return http.StatusConflict
	case errors.Is(err, domain.ErrCyclicCategory):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidPassword):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidRole):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrTenantNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrSameMonth):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
