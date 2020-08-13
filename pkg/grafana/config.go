package grafana

import (
	"fmt"
	"net/url"
	"os"
	"path"
)

func getGrafanaHost() (*url.URL, error) {
	if grafanaURL, exists := os.LookupEnv("GRAFANA_URL"); exists {
		u, err := url.Parse(grafanaURL)
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	return nil, fmt.Errorf("Require GRAFANA_URL (optionally GRAFANA_TOKEN & GRAFANA_USER")
}

func getGrafanaURL(urlPath string) (string, error) {
	grafanaHost, err := getGrafanaHost()
	if err != nil {
		return "", err
	}
	grafanaHost.Path = path.Join(grafanaHost.Path, urlPath)
	if token, exists := os.LookupEnv("GRAFANA_TOKEN"); exists {
		user, exists := os.LookupEnv("GRAFANA_USER")
		if !exists {
			user = "api_key"
		}
		grafanaHost.User = url.UserPassword(user, token)
	}
	return grafanaHost.String(), nil
}
