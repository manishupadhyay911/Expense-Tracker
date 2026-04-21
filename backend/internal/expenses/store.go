package expenses

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidIdempotencyKey = errors.New("idempotency key is required")
	ErrIdempotencyConflict   = errors.New("idempotency key was already used with different data")
)

type Store struct {
	mu          sync.RWMutex
	nextID      int64
	expenses    []Expense
	idempotency map[string]storedRequest
}

type storedRequest struct {
	fingerprint string
	expense     Expense
}

func NewStore() *Store {
	return &Store{
		nextID:      1,
		expenses:    make([]Expense, 0),
		idempotency: make(map[string]storedRequest),
	}
}

func (s *Store) Create(_ context.Context, req CreateExpenseRequest, idempotencyKey string) (Expense, bool, error) {
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return Expense{}, false, ErrInvalidIdempotencyKey
	}

	fingerprint := fingerprintFor(req)

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.idempotency[key]; ok {
		if existing.fingerprint != fingerprint {
			return Expense{}, false, ErrIdempotencyConflict
		}
		return existing.expense, false, nil
	}

	expense := Expense{
		ID:          s.nextID,
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
		CreatedAt:   time.Now().UTC().Round(time.Microsecond),
	}
	s.nextID++
	s.expenses = append(s.expenses, expense)
	s.idempotency[key] = storedRequest{
		fingerprint: fingerprint,
		expense:     expense,
	}

	return expense, true, nil
}

func (s *Store) List(category string, sortBy string) ([]Expense, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sortBy != "" && sortBy != "date_desc" && sortBy != "date_asc" {
		return nil, fmt.Errorf("unsupported sort value: %s", sortBy)
	}

	filter := strings.TrimSpace(category)
	items := make([]Expense, 0, len(s.expenses))
	for _, expense := range s.expenses {
		if filter != "" && expense.Category != filter {
			continue
		}
		items = append(items, expense)
	}

	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if !left.Date.Time.Equal(right.Date.Time) {
			if sortBy == "date_asc" {
				return left.Date.Time.Before(right.Date.Time)
			}
			return left.Date.Time.After(right.Date.Time)
		}
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}
		return left.ID > right.ID
	})

	return items, nil
}

func fingerprintFor(req CreateExpenseRequest) string {
	return fmt.Sprintf("%s|%s|%s|%s",
		req.Amount.String(),
		strings.TrimSpace(req.Category),
		strings.TrimSpace(req.Description),
		req.Date.String(),
	)
}
