CREATE TABLE kafka_outbox (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic TEXT NOT NULL,
    message_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    published_at TIMESTAMPTZ,
    UNIQUE (topic, message_key)
);
CREATE INDEX kafka_outbox_pending ON kafka_outbox (id) WHERE published_at IS NULL;
