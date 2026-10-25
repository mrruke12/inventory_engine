package order

import "errors"

var (
	ErrOrderExists error = errors.New("order already exists")
)
