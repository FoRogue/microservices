package eventing

import (
	"encoding/json"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestDecodeEnforcesTopicAndAggregateKey(t *testing.T) {
	const orderID = "c6a7cba8-06a5-475c-beb5-05c14abaec93"
	event := New(PaymentRequested, orderID, Payload{AmountCents: 2990})
	value, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	message := kafka.Message{Topic: PaymentsRequested, Key: []byte(orderID), Value: value}
	decoded, err := decode(message)
	if err != nil || decoded.ID != event.ID || !decoded.OccurredAt.Equal(event.OccurredAt) {
		t.Fatalf("decode returned %+v, %v", decoded, err)
	}
	message.Topic = PaymentsCompleted
	if _, err := decode(message); err == nil {
		t.Fatal("accepted payment request on the results topic")
	}
	message.Topic = PaymentsRequested
	message.Key = []byte("another-order")
	if _, err := decode(message); err == nil {
		t.Fatal("accepted a mismatched order key")
	}
}
