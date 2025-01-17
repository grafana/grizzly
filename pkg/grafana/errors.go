package grafana

import (
	"fmt"
	"strings"
)

type ErrUIDNameMismatch struct {
	UID  string
	Name string
}

func (e ErrUIDNameMismatch) Error() string {
	return fmt.Sprintf("uid '%s' and name '%s', don't match", e.UID, e.Name)
}

// ErrUidsMissing reports UIDs are missing for Dashboards
type ErrUidsMissing []string

func (e ErrUidsMissing) Error() string {
	return fmt.Sprintf("One or more dashboards have no UID set. UIDs are required for Grizzly to operate properly:\n - %s", strings.Join(e, "\n - "))
}

type APIResponse interface {
	Code() int
	Error() string
	String() string
}
