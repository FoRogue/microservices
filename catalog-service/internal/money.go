package internal

import (
	"encoding/json"
	"errors"
	"strings"
)

// Money holds the original decimal spelling until PostgreSQL applies NUMERIC(10,2).
type Money string

func (m *Money) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if len(data) == 0 || (data[0] != '-' && (data[0] < '0' || data[0] > '9')) {
		return errors.New("money must be a JSON number")
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	*m = Money(number)
	return nil
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.Number(m.databaseValue()))
}

func (m Money) databaseValue() string {
	if m == "" {
		return "0"
	}
	return string(m)
}

func (m Money) sign() int {
	mantissa := string(m)
	if i := strings.IndexAny(mantissa, "eE"); i >= 0 {
		mantissa = mantissa[:i]
	}
	if strings.Trim(mantissa, "-.0") == "" {
		return 0
	}
	if strings.HasPrefix(mantissa, "-") {
		return -1
	}
	return 1
}
