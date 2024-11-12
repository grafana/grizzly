package httputils

import (
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

type LoggedHTTPRoundTripper struct {
	DecoratedTransport http.RoundTripper
}

func (rt LoggedHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport
	if rt.DecoratedTransport != nil {
		transport = rt.DecoratedTransport
	}

	reqStr, _ := httputil.DumpRequest(req, true)
	log.Traceln(string(reqStr))

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	respStr, _ := httputil.DumpResponse(resp, true)
	log.Traceln(string(respStr))

	return resp, err
}
