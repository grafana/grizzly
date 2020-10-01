package grafana

import (
	"fmt"
	"strings"
)

// APIErr encapsulates an error from the Grafana API
type APIErr struct {
	err  error
	body []byte
}

func (e APIErr) Error() string {
	return fmt.Sprintf("Failed to parse Grafana response: %s.\n\nResponse:\n%s", e.err, string(e.body))
}

// ErrUidsMissing reports UIDs are missing for Dashboards
type ErrUidsMissing []string

func (e ErrUidsMissing) Error() string {
	return fmt.Sprintf("One or more dashboards have no UID set. UIDs are required for Grizzly to operate properly:\n - %s", strings.Join(e, "\n - "))
}
