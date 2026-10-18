package payment

type Status string

const (
	StatusCreating Status = "creating"
	StatusPending  Status = "pending"
	StatusPaid     Status = "paid"
	StatusExpired  Status = "expired"
	StatusFailed   Status = "failed"
)
