package expenses

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errInvalidMoney = errors.New("amount must be a valid non-negative currency value")

type Money struct {
	minor int64
}

func NewMoney(minor int64) Money {
	return Money{minor: minor}
}

func (m Money) Minor() int64 {
	return m.minor
}

func (m Money) String() string {
	sign := ""
	value := m.minor
	if value < 0 {
		sign = "-"
		value = -value
	}

	return fmt.Sprintf("%s%d.%02d", sign, value/100, value%100)
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *Money) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "null" {
		*m = Money{}
		return nil
	}

	if len(raw) > 0 && raw[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		money, err := parseMoney(text)
		if err != nil {
			return err
		}
		*m = money
		return nil
	}

	money, err := parseMoney(raw)
	if err != nil {
		return err
	}
	*m = money
	return nil
}

func parseMoney(raw string) (Money, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return Money{}, errInvalidMoney
	}
	if strings.HasPrefix(text, "-") {
		return Money{}, errInvalidMoney
	}
	if strings.HasPrefix(text, "+") {
		text = strings.TrimPrefix(text, "+")
	}

	parts := strings.SplitN(text, ".", 3)
	if len(parts) > 2 {
		return Money{}, errInvalidMoney
	}

	whole := parts[0]
	if whole == "" {
		return Money{}, errInvalidMoney
	}
	for _, r := range whole {
		if r < '0' || r > '9' {
			return Money{}, errInvalidMoney
		}
	}

	wholeValue, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return Money{}, errInvalidMoney
	}

	minor := wholeValue * 100
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) > 2 {
			return Money{}, errInvalidMoney
		}
		for _, r := range fraction {
			if r < '0' || r > '9' {
				return Money{}, errInvalidMoney
			}
		}
		for len(fraction) < 2 {
			fraction += "0"
		}
		fractionValue, err := strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return Money{}, errInvalidMoney
		}
		minor += fractionValue
	}

	return Money{minor: minor}, nil
}

