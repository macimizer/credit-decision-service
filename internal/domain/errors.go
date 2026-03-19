package domain

import "errors"

var (
	ErrNotFound    = errors.New("resource not found")
	ErrValidation  = errors.New("validation failed")
	ErrUnavailable = errors.New("service unavailable")
)
