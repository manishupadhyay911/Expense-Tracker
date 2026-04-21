package expenses

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	store *Store
	mux   *http.ServeMux
}

func NewHandler(store *Store) http.Handler {
	h := &Handler{
		store: store,
		mux:   http.NewServeMux(),
	}

	h.mux.HandleFunc("GET /health", h.handleHealth)
	h.mux.HandleFunc("POST /expenses", h.handleCreateExpense)
	h.mux.HandleFunc("GET /expenses", h.handleListExpenses)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/") {
		r = r.Clone(r.Context())
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
	}

	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	var req CreateExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid expense JSON")
		return
	}

	if err := validateCreateExpenseRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	expense, created, err := h.store.Create(r.Context(), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidIdempotencyKey):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrIdempotencyConflict):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create expense")
		}
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, expense)
}

func (h *Handler) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	sortBy := r.URL.Query().Get("sort")

	expenses, err := h.store.List(category, sortBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, expenses)
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	return nil
}

func validateCreateExpenseRequest(req CreateExpenseRequest) error {
	if req.Amount.Minor() <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if strings.TrimSpace(req.Category) == "" {
		return errors.New("category is required")
	}

	if strings.TrimSpace(req.Description) == "" {
		return errors.New("description is required")
	}

	if req.Date.Time.IsZero() {
		return errors.New("date is required")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
