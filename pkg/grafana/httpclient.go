package grafana

import (
	"net/http"
	"net/url"
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
	if httpProxy := os.Getenv("HTTP_PROXY"); httpProxy != "" {
		proxyUrl, err := url.Parse(httpProxy)
	Transport:
		&http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	return &http.Client{Timeout: timeout}, nil
}
