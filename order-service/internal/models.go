package internal

import (
	"fmt"
	"math/big"
	"time"
)

const (
	pending = "PENDING_PAYMENT"
	paid    = "PAID"
	failed  = "PAYMENT_FAILED"

	// Match payment-service's NUMERIC(10,2) limit.
	maxCents = 9_999_999_999
)

// Money stores an exact amount in cents and exposes a JSON number.
type Money int64

func (m Money) MarshalJSON() ([]byte, error) {
	// Orders only contain non-negative amounts.
	return fmt.Appendf(nil, "%d.%02d", m/100, m%100), nil
}

func priceInCents(value string) (Money, error) {
	// Catalog prices are PostgreSQL NUMERIC(10,2). Bound the input before parsing.
	if len(value) > 12 {
		return 0, errDependency
	}
	for _, ch := range value {
		if ch != '.' && (ch < '0' || ch > '9') {
			return 0, errDependency
		}
	}
	price, ok := new(big.Rat).SetString(value)
	if !ok {
		return 0, errDependency
	}
	price.Mul(price, big.NewRat(100, 1))
	if !price.IsInt() || !price.Num().IsInt64() || price.Num().Int64() > maxCents {
		return 0, errDependency
	}
	return Money(price.Num().Int64()), nil
}

type ItemRequest struct {
	ProductID int64 `json:"product_id" example:"1"`
	Quantity  int   `json:"quantity" example:"2"`
}

type CreateOrderRequest struct {
	Items []ItemRequest `json:"items"`
}

type OrderItem struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	UnitPrice   Money  `json:"unit_price" swaggertype:"number"`
}

type Order struct {
	ID        string      `json:"id"`
	UserID    int64       `json:"user_id"`
	Status    string      `json:"status" enums:"PENDING_PAYMENT,PAID,PAYMENT_FAILED"`
	Total     Money       `json:"total" swaggertype:"number"`
	Items     []OrderItem `json:"items"`
	PaymentID *int64      `json:"payment_id,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
