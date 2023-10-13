package grafana

import (
	"fmt"
	"net/url"
	"os"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
)

func GetClient() (*gclient.GrafanaHTTPAPI, error) {
	grafanaURL, exists := os.LookupEnv("GRAFANA_URL")
	if !exists {
		return nil, fmt.Errorf("require GRAFANA_URL (optionally GRAFANA_TOKEN & GRAFANA_USER)")
	}
	parsedUrl, err := url.Parse(grafanaURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Grafana URL")
	}

	transportConfig := gclient.DefaultTransportConfig().WithHost(parsedUrl.Host).WithSchemes([]string{parsedUrl.Scheme})
	if token, exists := os.LookupEnv("GRAFANA_TOKEN"); exists {
		if user, exists := os.LookupEnv("GRAFANA_USER"); exists {
			transportConfig.BasicAuth = url.UserPassword(user, token)
		} else {
			transportConfig.APIKey = token
		}
	}
	grafanaClient := gclient.NewHTTPClientWithConfig(nil, transportConfig)
	return grafanaClient, nil
}
