package handler

import (
	"errors"
	"net/http"

	"github.com/dcunha/finance/backend/internal/domain"
)

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
	case errors.Is(err, domain.ErrAlreadyPaused):
		return http.StatusConflict
	case errors.Is(err, domain.ErrAlreadyActive):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidFrequency):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
