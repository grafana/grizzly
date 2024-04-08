package mimir

import "fmt"

type ErrNoBinarySet struct {
	name string
}

func (e ErrNoBinarySet) Error() string {
	return fmt.Sprintf("binary %s isn't set. Install the binary or set the path using `mimir.%s-path`", e.name, e.name)
}
