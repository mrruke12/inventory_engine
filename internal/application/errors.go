package application

import "errors"

var (
	ErrIdempotencyKeyReused error = errors.New("idempotency key reused with different request")
)
