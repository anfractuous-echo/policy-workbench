package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidRequest    = errors.New("invalid request")
	ErrNotFound          = errors.New("not found")
	ErrActionUnavailable = errors.New("action not available")
)

func InvalidRequest(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
}
