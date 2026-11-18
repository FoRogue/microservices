package internal

import "time"

const (
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

// Payment is a single payment attempt for an order.
type Payment struct {
	ID        int64     `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    Money     `json:"amount" swaggertype:"number"`
	Status    string    `json:"status" example:"succeeded"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePaymentRequest is the payload for POST /api/payments.
type CreatePaymentRequest struct {
	OrderID string `json:"order_id" example:"order-123"`
	Amount  Money  `json:"amount" swaggertype:"number" example:"1999.90"`
}

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"payment not found"`
}
