package grafana

import (
	"fmt"
	"strings"
)

// ErrUidsMissing reports UIDs are missing for Dashboards
type ErrUidsMissing []string

func (e ErrUidsMissing) Error() string {
	return fmt.Sprintf("One or more dashboards have no UID set. UIDs are required for Grizzly to operate properly:\n - %s", strings.Join(e, "\n - "))
}
