package entity

type DashboardSummary struct {
	TotalIncome     float64 `json:"total_income"`
	TotalExpenses   float64 `json:"total_expenses"`
	Balance         float64 `json:"balance"`
	PreviousBalance float64 `json:"previous_balance"`
	IncomeCount     int     `json:"income_count"`
	ExpenseCount    int     `json:"expense_count"`
}

type CategoryTotal struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
}
