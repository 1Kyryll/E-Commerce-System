CREATE TABLE carts (
    id         UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
    id         UUID        PRIMARY KEY,
    cart_id    UUID        NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id UUID        NOT NULL REFERENCES products(id),
    quantity   INTEGER     NOT NULL CHECK (quantity > 0),
    added_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (cart_id, product_id)
);

CREATE INDEX idx_cart_items_cart_id ON cart_items (cart_id);
