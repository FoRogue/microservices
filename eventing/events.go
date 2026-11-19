package eventing

import (
	"crypto/rand"
	"os"
	"strings"
	"time"
)

const (
	PaymentsRequested = "payments.requested"
	PaymentsCompleted = "payments.completed"
	PaymentRequested  = "PaymentRequested"
	PaymentSucceeded  = "PaymentSucceeded"
	PaymentFailed     = "PaymentFailed"
	PaymentGroup      = "payment-workers"
	OrderGroup        = "order-payment-results"
	MaxAmountCents    = 9_999_999_999
)

type Payload struct {
	AmountCents int64 `json:"amountCents"`
	PaymentID   int64 `json:"paymentId,omitempty"`
}

type Event struct {
	ID          string    `json:"eventId"`
	Type        string    `json:"eventType"`
	OccurredAt  time.Time `json:"occurredAt"`
	AggregateID string    `json:"aggregateId"`
	Payload     Payload   `json:"payload"`
}

func New(kind, orderID string, payload Payload) Event {
	return Event{ID: rand.Text(), Type: kind, OccurredAt: time.Now().UTC(), AggregateID: orderID, Payload: payload}
}

func Brokers() []string {
	if value := os.Getenv("KAFKA_BROKERS"); value != "" {
		return strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	}
	return []string{"localhost:9092"}
}
