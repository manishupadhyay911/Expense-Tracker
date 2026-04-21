package handler

import (
	"net/http"

	"expense-tracker/backend/internal/expenses"
)

var defaultHandler = expenses.NewHandler(expenses.NewStore())

func Handler(w http.ResponseWriter, r *http.Request) {
	defaultHandler.ServeHTTP(w, r)
}
