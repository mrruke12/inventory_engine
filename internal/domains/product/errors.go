package product

import "errors"

var (
	ErrorNotEnoughProduct error = errors.New("not enough product quantity")
)
