package expenses

import "time"

type Expense struct {
	ID          int64       `json:"id"`
	Amount      Money       `json:"amount"`
	Category    string      `json:"category"`
	Description string      `json:"description"`
	Date        ExpenseDate `json:"date"`
	CreatedAt   time.Time   `json:"created_at"`
}

type CreateExpenseRequest struct {
	Amount      Money       `json:"amount"`
	Category    string      `json:"category"`
	Description string      `json:"description"`
	Date        ExpenseDate `json:"date"`
}

