package entity

import (
	"time"

	"github.com/google/uuid"
)

type Membership struct {
	ID           uuid.UUID `json:"id"`
	GlobalUserID uuid.UUID `json:"global_user_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	SchemaUserID uuid.UUID `json:"schema_user_id"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TenantMembership is used in the tenant selector after login.
type TenantMembership struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Role       string    `json:"role"`
}
