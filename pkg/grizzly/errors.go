package grizzly

import (
	"errors"
	"fmt"
)

// ErrNotFound is used to signal a missing resource
var ErrNotFound = errors.New("not found")

// ErrNotImplemented signals a feature that is not supported by a provider
var ErrNotImplemented = errors.New("not implemented")

// APIErr encapsulates an error from the Grafana API
type APIErr struct {
	Err  error
	Body []byte
}

func (e APIErr) Error() string {
	return fmt.Sprintf("Failed to parse Grafana response: %s.\n\nResponse:\n%s", e.Err, string(e.Body))
}
