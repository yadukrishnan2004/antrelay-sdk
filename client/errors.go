package client

import "fmt"

// ValidationError is returned when a Config field fails validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("antrelay: invalid field %q: %s", e.Field, e.Message)
}