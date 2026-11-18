CREATE TABLE IF NOT EXISTS payments (
	id BIGSERIAL PRIMARY KEY,
	order_id TEXT NOT NULL,
	amount NUMERIC(10,2) NOT NULL,
	status TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);
