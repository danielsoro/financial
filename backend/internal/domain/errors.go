package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrDuplicateEmail     = errors.New("email already registered")
	ErrDuplicateCategory  = errors.New("category name already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrForbidden          = errors.New("you do not have permission to perform this action")
	ErrCategoryInUse      = errors.New("category is in use by transactions")
	ErrDuplicateLimit     = errors.New("expense limit already exists for this period")
	ErrCyclicCategory     = errors.New("cannot create cyclic category hierarchy")
	ErrInvalidPassword    = errors.New("current password is incorrect")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrDuplicateDomain    = errors.New("domain already in use")
	ErrInvalidRole        = errors.New("invalid role")
	ErrSameMonth          = errors.New("source and target month must be different")
	ErrAlreadyPaused      = errors.New("recurring transaction is already paused")
	ErrAlreadyActive      = errors.New("recurring transaction is already active")
	ErrInvalidFrequency   = errors.New("invalid frequency")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrAlreadyMember      = errors.New("user is already a member of this tenant")
	ErrMaxTenantsReached  = errors.New("maximum number of owned tenants reached")
	ErrInviteExpired      = errors.New("invite has expired")
	ErrInviteAlreadyUsed  = errors.New("invite has already been accepted")
	ErrNoMemberships      = errors.New("user has no tenant memberships")
	ErrDuplicateTenant    = errors.New("tenant name already in use")
)
