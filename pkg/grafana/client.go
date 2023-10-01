package grafana

import (
	"fmt"
	"net/url"
	"os"

	grafana "github.com/grafana/grafana-api-golang-client"
)

func getClient() (*grafana.Client, error) {
	grafanaURL, exists := os.LookupEnv("GRAFANA_URL")
	if !exists {
		return nil, fmt.Errorf("require GRAFANA_URL (optionally GRAFANA_TOKEN & GRAFANA_USER")
	}
	cfg := grafana.Config{}

	if token, exists := os.LookupEnv("GRAFANA_TOKEN"); exists {
		if user, exists := os.LookupEnv("GRAFANA_USER"); exists {
			cfg.BasicAuth = url.UserPassword(user, token)
		} else {
			cfg.APIKey = token
		}
	}

	return grafana.New(grafanaURL, cfg)
}
