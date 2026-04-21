package expenses

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var errInvalidDate = errors.New("date must be a valid YYYY-MM-DD value")

type ExpenseDate struct {
	time.Time
}

func ParseExpenseDate(raw string) (ExpenseDate, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ExpenseDate{}, errInvalidDate
	}

	parsed, err := time.Parse("2006-01-02", text)
	if err != nil {
		return ExpenseDate{}, errInvalidDate
	}
	return ExpenseDate{Time: parsed.UTC()}, nil
}

func (d ExpenseDate) String() string {
	return d.Time.UTC().Format("2006-01-02")
}

func (d ExpenseDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *ExpenseDate) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return errInvalidDate
	}

	parsed, err := ParseExpenseDate(text)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

