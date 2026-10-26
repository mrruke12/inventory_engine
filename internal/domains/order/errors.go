package order

import "errors"

var (
	ErrOrderExists    error = errors.New("order already exists")
	ErrOrderNotFound  error = errors.New("order is not found")
	ErrOrderCanceled  error = errors.New("order is canceled")
	ErrOrderCompleted error = errors.New("order is completed")
)
