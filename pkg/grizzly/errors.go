package grizzly

import (
	"errors"
)

// ErrNotFound is used to signal a missing resource
var ErrNotFound = errors.New("not found")
