package grafana

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

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

func writeOrLog(w http.ResponseWriter, content []byte) {
	if _, err := w.Write(content); err != nil {
		log.Errorf("error writing response: %v", err)
	}
}
