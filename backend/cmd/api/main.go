package main

import (
	"log"
	"net/http"
	"os"

	"expense-tracker/backend/internal/expenses"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := expenses.NewHandler(expenses.NewStore())

	addr := ":" + port
	log.Printf("backend listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
