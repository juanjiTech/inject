package inject

import "errors"

var (
	ErrValueNotFound  = errors.New("value not found")
	ErrValueCanNotSet = errors.New("value can not set")
)
