package domain

import "time"

type Return struct {
	ID          int       `db:"id"`
	OrderID     string    `db:"order_id"`
	RecipientID string    `db:"recipient_id"`
	ReturnDate  time.Time `db:"return_date"`
}
