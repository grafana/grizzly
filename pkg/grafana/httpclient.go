package grafana

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

func NewHttpClient() (*http.Client, error) {
	timeout := 10 * time.Second
	if timeoutStr := os.Getenv("GRIZZLY_HTTP_TIMEOUT"); timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, err
		}
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	return &http.Client{Timeout: timeout}, nil
}
