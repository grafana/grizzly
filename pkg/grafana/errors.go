package grafana

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// ErrUidsMissing reports UIDs are missing for Dashboards
type ErrUidsMissing []string

func (e ErrUidsMissing) Error() string {
	return fmt.Sprintf("One or more dashboards have no UID set. UIDs are required for Grizzly to operate properly:\n - %s", strings.Join(e, "\n - "))
}

type ErrNon200Response struct {
	Type     string
	UID      string
	Response *http.Response
}

func NewErrNon200Response(typ, uid string, resp *http.Response) ErrNon200Response {
	return ErrNon200Response{
		Type:     typ,
		UID:      uid,
		Response: resp,
	}
}
func (e ErrNon200Response) Error() string {
	body, _ := ioutil.ReadAll(e.Response.Body)
	status := e.Response.Status
	return fmt.Sprintf("Non-200 response from Grafana while applying %s %s: %s %s", e.Type, e.UID, status, string(body))
}
