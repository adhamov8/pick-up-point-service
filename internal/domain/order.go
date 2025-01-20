package domain

import (
	"database/sql"
	"time"
)

type Order struct {
	OrderID       string       `db:"order_id"`
	RecipientID   string       `db:"recipient_id"`
	ExpiryDate    time.Time    `db:"expiry_date"`
	Status        string       `db:"status"`
	DeliveryDate  sql.NullTime `db:"delivery_date"`
	ReturnDate    sql.NullTime `db:"return_date"`
	Weight        float32      `db:"weight"`
	Cost          float32      `db:"cost"`
	PackagingType string       `db:"packaging_type"`
}
