package entity

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	UserID       uuid.UUID `json:"user_id"`
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Type         string    `json:"type"`
	Amount       float64   `json:"amount"`
	Description  string    `json:"description"`
	Date         string    `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TransactionFilter struct {
	TenantID   uuid.UUID
	Type       string
	CategoryID *uuid.UUID
	StartDate  string
	EndDate    string
	Page       int
	PerPage    int
}

type PaginatedTransactions struct {
	Data       []Transaction `json:"data"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}
