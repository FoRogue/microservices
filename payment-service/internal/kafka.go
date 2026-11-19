package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ForRogue/microservices/eventing"
)

func (repo *Repository) ApplyPaymentRequest(ctx context.Context, event eventing.Event) (string, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	cents := event.Payload.AmountCents
	amount := Money(fmt.Sprintf("%d.%02d", cents/100, cents%100))
	req := CreatePaymentRequest{OrderID: event.AggregateID, Amount: amount}
	key := "kafka-order:" + event.AggregateID
	payment, err := createPayment(ctx, tx, req, paymentStatus(amount), key)
	if errors.Is(err, sql.ErrNoRows) {
		return "CONFLICT", nil
	}
	if err != nil {
		return "", err
	}
	kind := eventing.PaymentSucceeded
	if payment.Status == StatusFailed {
		kind = eventing.PaymentFailed
	}
	result := eventing.New(kind, event.AggregateID, eventing.Payload{AmountCents: cents, PaymentID: payment.ID})
	if err := eventing.Enqueue(ctx, tx, eventing.PaymentsCompleted, result); err != nil {
		return "", err
	}
	return payment.Status, tx.Commit()
}
