-- +goose Up
CREATE TABLE IF NOT EXISTS returns (
    return_id TEXT PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    return_date DATE NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
    );

-- +goose Down
DROP TABLE IF EXISTS returns;