package entity

import (
	"time"

	"github.com/google/uuid"
)

type GlobalUser struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	EmailVerified       bool       `json:"email_verified"`
	EmailToken          *string    `json:"-"`
	EmailTokenExpiresAt *time.Time `json:"-"`
	MaxOwnedTenants     int        `json:"max_owned_tenants"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
