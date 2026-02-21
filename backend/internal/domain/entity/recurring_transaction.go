package entity

import (
	"time"

	"github.com/google/uuid"
)

type DeleteMode string

const (
	DeleteModeAll              DeleteMode = "all"
	DeleteModeFutureAndCurrent DeleteMode = "future_and_current"
	DeleteModeFutureOnly       DeleteMode = "future_only"
)

type RecurringTransaction struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	CategoryID     uuid.UUID  `json:"category_id"`
	CategoryName   string     `json:"category_name,omitempty"`
	Type           string     `json:"type"`
	Amount         float64    `json:"amount"`
	Description    string     `json:"description"`
	Frequency      string     `json:"frequency"`
	StartDate      string     `json:"start_date"`
	EndDate        *string    `json:"end_date"`
	MaxOccurrences *int       `json:"max_occurrences"`
	DayOfMonth     *int       `json:"day_of_month"`
	IsActive       bool       `json:"is_active"`
	PausedAt       *time.Time `json:"paused_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type RecurringTransactionFilter struct {
	Type     string
	IsActive *bool
	Page     int
	PerPage  int
}

type PaginatedRecurringTransactions struct {
	Data       []RecurringTransaction `json:"data"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	PerPage    int                    `json:"per_page"`
	TotalPages int                    `json:"total_pages"`
}
