package domain

import (
	"errors"
	"fmt"
)

var ErrBatchNotFound = errors.New("batch not found")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}
