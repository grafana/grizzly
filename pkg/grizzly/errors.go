package grizzly

import (
	"errors"
)

// ErrNotFound is used to signal a missing resource
var ErrNotFound = errors.New("not found")

// ErrNotImplemented signals a feature that is not supported by a provider
var ErrNotImplemented = errors.New("not implemented")
