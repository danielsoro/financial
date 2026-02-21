package entity

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Domain     *string    `json:"domain"`
	SchemaName string     `json:"schema_name"`
	IsActive   bool       `json:"is_active"`
	OwnerID    *uuid.UUID `json:"owner_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
