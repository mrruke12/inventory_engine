package env

import "errors"

var (
	ErrNotPointerToStruct   error = errors.New("is not pointer to struct")
	ErrUnsupportedFieldType error = errors.New("unsupported field type")
)
