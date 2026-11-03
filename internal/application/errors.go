package application

import "errors"

var (
	ErrIdempotencyKeyReused error = errors.New("idempotency key reused with different request")
	ErrImpossibleCondition  error = errors.New("impossible condition")
	ErrMustUpdate           error = errors.New("order supposed to be updated, but did not change")
	ErrMustNotUpdate        error = errors.New("order supposed to be not updated, but did change")
)
