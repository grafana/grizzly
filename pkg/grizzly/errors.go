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

type UnrecognisedFormatError struct {
	File string
}

func (e UnrecognisedFormatError) Error() string {
	return fmt.Sprintf("unrecognized format for %s", e.File)
}

func NewUnrecognisedFormatError(file string) UnrecognisedFormatError {
	return UnrecognisedFormatError{
		File: file,
	}
}

type Warning struct {
	Err error
}

func NewWarning(err error) Warning {
	return Warning{err}
}

func (w Warning) Error() string {
	return w.Err.Error()
}

func IsWarning(err any) bool {
	_, ok := err.(Warning)
	return ok
}
