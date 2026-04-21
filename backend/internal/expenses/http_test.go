package expenses

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateExpenseSuccessAndIdempotency(t *testing.T) {
	handler := NewHandler(NewStore())

	first := performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"12.50","category":"Food","description":"Lunch","date":"2026-04-21"}`, "idem-1")
	if first.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, first.Code)
	}

	var created Expense
	if err := json.Unmarshal(first.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected id 1, got %d", created.ID)
	}
	if created.Amount.String() != "12.50" {
		t.Fatalf("expected amount 12.50, got %s", created.Amount.String())
	}

	second := performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"12.50","category":"Food","description":"Lunch","date":"2026-04-21"}`, "idem-1")
	if second.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, second.Code)
	}

	var repeated Expense
	if err := json.Unmarshal(second.Body.Bytes(), &repeated); err != nil {
		t.Fatalf("unmarshal repeated response: %v", err)
	}
	if repeated.ID != created.ID {
		t.Fatalf("expected repeated id %d, got %d", created.ID, repeated.ID)
	}
}

func TestCreateExpenseValidation(t *testing.T) {
	handler := NewHandler(NewStore())

	tests := []struct {
		name       string
		body       string
		statusCode int
	}{
		{name: "missing amount", body: `{"category":"Food","description":"Lunch","date":"2026-04-21"}`, statusCode: http.StatusBadRequest},
		{name: "bad date", body: `{"amount":"12.50","category":"Food","description":"Lunch","date":"2026-99-21"}`, statusCode: http.StatusBadRequest},
		{name: "missing key", body: `{"amount":"12.50","category":"Food","description":"Lunch","date":"2026-04-21"}`, statusCode: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString(tc.body))
			if tc.name != "missing key" {
				req.Header.Set("Idempotency-Key", "idem-validate")
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.statusCode {
				t.Fatalf("expected status %d, got %d", tc.statusCode, rr.Code)
			}
		})
	}
}

func TestListExpensesFilteringAndSort(t *testing.T) {
	handler := NewHandler(NewStore())

	performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"10.00","category":"Food","description":"Coffee","date":"2026-04-20"}`, "idem-1")
	performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"20.00","category":"Travel","description":"Cab","date":"2026-04-21"}`, "idem-2")
	performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"30.00","category":"Food","description":"Dinner","date":"2026-04-22"}`, "idem-3")

	all := performRequest(t, handler, http.MethodGet, "/expenses?sort=date_desc", "", "")
	if all.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, all.Code)
	}

	var expenses []Expense
	if err := json.Unmarshal(all.Body.Bytes(), &expenses); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(expenses) != 3 {
		t.Fatalf("expected 3 expenses, got %d", len(expenses))
	}
	if expenses[0].Description != "Dinner" {
		t.Fatalf("expected newest expense first, got %s", expenses[0].Description)
	}

	filtered := performRequest(t, handler, http.MethodGet, "/expenses?category=Food&sort=date_desc", "", "")
	if filtered.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, filtered.Code)
	}
	if err := json.Unmarshal(filtered.Body.Bytes(), &expenses); err != nil {
		t.Fatalf("unmarshal filtered list: %v", err)
	}
	if len(expenses) != 2 {
		t.Fatalf("expected 2 food expenses, got %d", len(expenses))
	}
	for _, expense := range expenses {
		if expense.Category != "Food" {
			t.Fatalf("unexpected category %s", expense.Category)
		}
	}
}

func TestListExpensesOldestFirstSort(t *testing.T) {
	handler := NewHandler(NewStore())

	performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"10.00","category":"Food","description":"Coffee","date":"2026-04-22"}`, "idem-oldest-1")
	performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"20.00","category":"Food","description":"Lunch","date":"2026-04-20"}`, "idem-oldest-2")

	resp := performRequest(t, handler, http.MethodGet, "/expenses?sort=date_asc", "", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var expenses []Expense
	if err := json.Unmarshal(resp.Body.Bytes(), &expenses); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(expenses))
	}
	if expenses[0].Description != "Lunch" {
		t.Fatalf("expected oldest expense first, got %s", expenses[0].Description)
	}
}

func TestListExpensesRejectsInvalidSort(t *testing.T) {
	handler := NewHandler(NewStore())

	resp := performRequest(t, handler, http.MethodGet, "/expenses?sort=bad_value", "", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestOptionsPreflightSetsCorsHeaders(t *testing.T) {
	handler := NewHandler(NewStore())

	req := httptest.NewRequest(http.MethodOptions, "/expenses", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected allow-origin header, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected allow-methods header to be set")
	}
}

func TestApiPrefixRoutesWork(t *testing.T) {
	handler := NewHandler(NewStore())

	resp := performRequest(t, handler, http.MethodGet, "/api/health", "", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestIdempotencyConflict(t *testing.T) {
	handler := NewHandler(NewStore())

	performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"10.00","category":"Food","description":"Coffee","date":"2026-04-20"}`, "idem-conflict")
	second := performRequest(t, handler, http.MethodPost, "/expenses", `{"amount":"15.00","category":"Food","description":"Coffee","date":"2026-04-20"}`, "idem-conflict")
	if second.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, second.Code)
	}
}

func performRequest(t *testing.T, handler http.Handler, method, target, body, idempotencyKey string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
