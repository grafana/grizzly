package dash

import (
	"net/url"
	"os"
)

// Config provides configuration to `grafana-dash`
type Config struct {
	GrafanaDir  string
	GrafanaURL  string
	JsonnetPath string
}

// ParseEnvironment parses necessary environment variables
func ParseEnvironment() (*Config, error) {
	var config Config
	if grafanaDir, exists := os.LookupEnv("GRAFANA_DIR"); exists {
		config.GrafanaDir = grafanaDir
	}
	if grafanaURL, exists := os.LookupEnv("GRAFANA_URL"); exists {
		u, err := url.Parse(grafanaURL)
		if err != nil {
			return nil, err
		}
		config.GrafanaURL = u.String()
		if token, exists := os.LookupEnv("GRAFANA_TOKEN"); exists {
			user, exists := os.LookupEnv("GRAFANA_USER")
			if !exists {
				user = "api_key"
			}
			u.User = url.UserPassword(user, token)
			config.GrafanaURL = u.String()
		}
	}
	return &config, nil
}
