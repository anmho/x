package domain

import "errors"

var (
	ErrMissingMessage = errors.New("message is required")
	ErrNotFound       = errors.New("agent run not found")
)
