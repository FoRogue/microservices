package eventing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func Enqueue(ctx context.Context, tx *sql.Tx, topic string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO kafka_outbox (topic, message_key, payload)
		VALUES ($1, $2, $3::text::jsonb)
		ON CONFLICT (topic, message_key) DO NOTHING`, topic, event.AggregateID, string(data))
	return err
}

func Publish(ctx context.Context, db *sql.DB, topic string) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(Brokers()...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		BatchSize:    1,
		MaxAttempts:  1,
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  3 * time.Second,
	}
	defer writer.Close()
	for ctx.Err() == nil {
		attempt, cancel := context.WithTimeout(ctx, 8*time.Second)
		found, err := publishOne(attempt, db, writer, topic)
		cancel()
		if err != nil && ctx.Err() == nil {
			log.Printf("outbox publish: %v", err)
		}
		if err != nil || !found {
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
		}
	}
}

func publishOne(ctx context.Context, db *sql.DB, writer *kafka.Writer, topic string) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var id int64
	var key string
	var value []byte
	err = tx.QueryRowContext(ctx, `
		SELECT id, message_key, payload::text FROM kafka_outbox
		WHERE topic = $1 AND published_at IS NULL ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED`, topic).
		Scan(&id, &key, &value)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := writer.WriteMessages(ctx, kafka.Message{Topic: topic, Key: []byte(key), Value: value}); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE kafka_outbox SET published_at = now() WHERE id = $1", id); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
