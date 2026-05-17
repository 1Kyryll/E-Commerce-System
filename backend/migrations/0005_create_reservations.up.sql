CREATE TABLE reservations (
    id              UUID        PRIMARY KEY,
    idempotency_key UUID        NOT NULL UNIQUE,
    product_id      UUID        NOT NULL REFERENCES products(id),
    user_id         UUID        NOT NULL REFERENCES users(id),
    quantity        INTEGER     NOT NULL CHECK (quantity > 0),
    status          VARCHAR(32) NOT NULL
        CHECK (status IN ('active', 'consumed', 'released')),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    consumed_at     TIMESTAMPTZ,
    released_at     TIMESTAMPTZ
);

CREATE INDEX idx_reservations_active_expiring
    ON reservations (expires_at) WHERE status = 'active';

CREATE INDEX idx_reservations_active_by_product
    ON reservations (product_id) WHERE status = 'active';
