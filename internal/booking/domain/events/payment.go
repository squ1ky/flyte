package events

type PaymentStatus string

const (
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

type PaymentRequestEvent struct {
	BookingID   string `json:"booking_id"`
	UserID      int64  `json:"user_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type PaymentResultEvent struct {
	BookingID    string        `json:"booking_id"`
	PaymentID    string        `json:"payment_id"`
	Status       PaymentStatus `json:"status"`
	ErrorMessage string        `json:"error_message,omitempty"`
	ProcessedAt  string        `json:"processed_at"`
}
