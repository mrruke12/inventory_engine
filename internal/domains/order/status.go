package order

type Status string

const (
	StatusPendingPayment Status = "pending"
	StatusExpired        Status = "expired"
	StatusCompleted      Status = "completed"
)
