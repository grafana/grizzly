package grizzly

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is used to signal a missing resource
	ErrNotFound = errors.New("not found")

	// ErrNotImplemented signals a feature that is not supported by a provider
	ErrNotImplemented = errors.New("not implemented")

	// ErrHandlerNotFound indicates that no handler was found for a particular resource Kind.
	ErrHandlerNotFound = errors.New("handler not found")
)

// APIErr encapsulates an error from the Grafana API
type APIErr struct {
	Err  error
	Body []byte
}

func (e APIErr) Error() string {
	return fmt.Sprintf("Failed to parse Grafana response: %s.\n\nResponse:\n%s", e.Err, string(e.Body))
}
