package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrDuplicateEmail    = errors.New("email already registered")
	ErrDuplicateCategory = errors.New("category name already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrForbidden         = errors.New("you do not have permission to perform this action")
	ErrCategoryInUse     = errors.New("category is in use by transactions")
	ErrDuplicateLimit    = errors.New("expense limit already exists for this period")
	ErrCyclicCategory    = errors.New("cannot create cyclic category hierarchy")
	ErrInvalidPassword   = errors.New("current password is incorrect")
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrDuplicateDomain   = errors.New("domain already in use")
	ErrInvalidRole       = errors.New("invalid role")
	ErrSameMonth         = errors.New("source and target month must be different")
)
