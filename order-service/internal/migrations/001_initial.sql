CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING_PAYMENT', 'PAID', 'PAYMENT_FAILED')),
    total_cents BIGINT NOT NULL CHECK (total_cents BETWEEN 0 AND 9999999999),
    payment_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((status = 'PENDING_PAYMENT') = (payment_id IS NULL))
);

CREATE TABLE order_items (
    order_id UUID NOT NULL REFERENCES orders (id),
    product_id BIGINT NOT NULL,
    product_name TEXT NOT NULL,
    quantity INT NOT NULL CHECK (quantity BETWEEN 1 AND 1000),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents BETWEEN 0 AND 9999999999),
    PRIMARY KEY (order_id, product_id)
);
