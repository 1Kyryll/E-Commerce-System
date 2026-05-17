CREATE TABLE orders (
    id              UUID           PRIMARY KEY,
    idempotency_key UUID           NOT NULL UNIQUE,
    user_id         UUID           NOT NULL REFERENCES users(id),
    total_amount    NUMERIC(18, 4) NOT NULL,
    total_currency  CHAR(3)        NOT NULL,
    status          VARCHAR(32)    NOT NULL
        CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_created ON orders (user_id, created_at DESC);

CREATE TABLE order_items (
    id                  UUID           PRIMARY KEY,
    order_id            UUID           NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id          UUID           NOT NULL REFERENCES products(id),
    reservation_id      UUID,
    quantity            INTEGER        NOT NULL CHECK (quantity > 0),
    unit_price_amount   NUMERIC(18, 4) NOT NULL,
    unit_price_currency CHAR(3)        NOT NULL
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
