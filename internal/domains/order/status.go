package order

type Status string

const (
	StatusPending   Status = "pending"   // A reserved order waiting to be completed/canceled
	StatusReleased  Status = "released"  // An order that wasn't completed and canceled manually
	StatusExpired   Status = "expired"   // An order that wasn't completed and canceled automatically
	StatusCompleted Status = "completed" // An order that was completed
	StatusCanceled  Status = "canceled"  // An order that was completed but cancelled after
)
