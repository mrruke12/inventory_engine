package order

type Status string

const (
	StatusPendingPayment Status = "pending_payment"
	StatusExpired        Status = "expired"
	StatusCompleted      Status = "completed"
)
