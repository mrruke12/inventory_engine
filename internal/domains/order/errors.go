package order

import "errors"

var (
	ErrOrderExists   error = errors.New("order already exists")
	ErrOrderNotFound error = errors.New("order is not found")

	ErrOrderPending   error = errors.New("order is pending")
	ErrOrderExpired   error = errors.New("order is expired")
	ErrOrderReleased  error = errors.New("order is released")
	ErrOrderCompleted error = errors.New("order is completed")
	ErrOrderCanceled  error = errors.New("order is canceled")
	ErrInvalidStatus  error = errors.New("invalid status")
)

func ErrOrderByStatus(status Status) error {
	switch status {
	case StatusPending:
		return ErrOrderPending
	case StatusExpired:
		return ErrOrderExpired
	case StatusReleased:
		return ErrOrderReleased
	case StatusCompleted:
		return ErrOrderCompleted
	case StatusCanceled:
		return ErrOrderCanceled
	default:
		return ErrInvalidStatus
	}
}
