package entity

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Token      string     `json:"-"`
	InvitedBy  uuid.UUID  `json:"invited_by"`
	AcceptedAt *time.Time `json:"accepted_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type InviteInfo struct {
	TenantName string `json:"tenant_name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	UserExists bool   `json:"user_exists"`
}
