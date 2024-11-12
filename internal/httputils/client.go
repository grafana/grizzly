package httputils

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

var defaultTimeout = 10 * time.Second

func NewHTTPClient() (*http.Client, error) {
	timeout := defaultTimeout

	// TODO: Move this configuration to the global configuration
	if timeoutStr := os.Getenv("GRIZZLY_HTTP_TIMEOUT"); timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, err
		}

		timeout = time.Duration(timeoutSeconds) * time.Second
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &LoggedHTTPRoundTripper{},
	}, nil
}
