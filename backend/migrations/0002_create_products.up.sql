CREATE TABLE products (
    id                  UUID           PRIMARY KEY,
    name                VARCHAR(255)   NOT NULL,
    price_amount        NUMERIC(18, 4) NOT NULL,
    price_currency      CHAR(3)        NOT NULL,
    inventory_total     INTEGER        NOT NULL CHECK (inventory_total >= 0),
    inventory_available INTEGER        NOT NULL CHECK (inventory_available >= 0),
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_created_at ON products (created_at DESC, id DESC);
