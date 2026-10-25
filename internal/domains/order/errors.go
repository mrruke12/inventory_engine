package order

import "errors"

var (
	ErrorOrderExists error = errors.New("order already exists")
)
