package product

import "errors"

var (
	ErrNotEnoughProduct error = errors.New("not enough product quantity")
)
