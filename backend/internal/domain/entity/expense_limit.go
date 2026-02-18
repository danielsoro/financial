package entity

import (
	"time"

	"github.com/google/uuid"
)

type ExpenseLimit struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	CategoryID   *uuid.UUID `json:"category_id"`
	CategoryName string     `json:"category_name,omitempty"`
	Month        int        `json:"month"`
	Year         int        `json:"year"`
	Amount       float64    `json:"amount"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type LimitProgress struct {
	Limit      ExpenseLimit `json:"limit"`
	Spent      float64      `json:"spent"`
	Remaining  float64      `json:"remaining"`
	Percentage float64      `json:"percentage"`
}
