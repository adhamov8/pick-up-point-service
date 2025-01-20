package events

type OrderEvent struct {
	OrderID     string      `json:"order_id"`
	Event       string      `json:"event"`
	Timestamp   string      `json:"timestamp"`
	Details     interface{} `json:"details,omitempty"`
	RecipientID string      `json:"recipient_id,omitempty"`
}
