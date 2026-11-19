package eventing

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/segmentio/kafka-go"
)

var orderIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func Consume(ctx context.Context, group, topic string, apply func(context.Context, Event) (string, error)) {
	for ctx.Err() == nil {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     Brokers(),
			GroupID:     group,
			Topic:       topic,
			StartOffset: kafka.FirstOffset,
			MaxWait:     time.Second,
		})
		err := consume(ctx, reader, apply)
		reader.Close()
		if ctx.Err() == nil {
			log.Printf("consumer reconnect: %v", err)
			waitRetry(ctx)
		}
	}
}

func consume(ctx context.Context, reader *kafka.Reader, apply func(context.Context, Event) (string, error)) error {
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		event, err := decode(msg)
		metadata := []any{msg.Topic, msg.Partition, msg.Offset, string(msg.Key)}
		if err != nil {
			log.Printf("topic=%s partition=%d offset=%d key=%q invalid=%v", append(metadata, err)...)
		} else {
			for {
				attempt, cancel := context.WithTimeout(ctx, 5*time.Second)
				status, applyErr := apply(attempt, event)
				cancel()
				if applyErr == nil {
					log.Printf("topic=%s partition=%d offset=%d key=%q eventId=%s status=%s",
						append(metadata, event.ID, status)...)
					break
				}
				log.Printf("topic=%s partition=%d offset=%d key=%q retry=%v", append(metadata, applyErr)...)
				if !waitRetry(ctx) {
					return ctx.Err()
				}
			}
		}
		attempt, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = reader.CommitMessages(attempt, msg)
		cancel()
		if err != nil {
			return err
		}
	}
}

func decode(msg kafka.Message) (Event, error) {
	var event Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return Event{}, err
	}
	if event.ID == "" || !orderIDPattern.MatchString(event.AggregateID) ||
		event.AggregateID != string(msg.Key) || event.OccurredAt.IsZero() ||
		event.Payload.AmountCents < 0 || event.Payload.AmountCents > MaxAmountCents {
		return Event{}, errors.New("invalid event envelope")
	}
	switch {
	case msg.Topic == PaymentsRequested && event.Type == PaymentRequested && event.Payload.PaymentID == 0:
	case msg.Topic == PaymentsCompleted && (event.Type == PaymentSucceeded || event.Type == PaymentFailed) && event.Payload.PaymentID > 0:
	default:
		return Event{}, errors.New("topic, event type, or payload identity mismatch")
	}
	return event, nil
}

func waitRetry(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Second):
		return true
	}
}
