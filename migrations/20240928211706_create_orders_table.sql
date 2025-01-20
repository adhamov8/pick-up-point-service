-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    order_id TEXT PRIMARY KEY,
    recipient_id TEXT NOT NULL,
    expiry_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL,
    delivery_date TIMESTAMP,
    return_date TIMESTAMP,
    weight REAL NOT NULL,
    cost REAL NOT NULL,
    packaging_type VARCHAR(50) NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS orders;