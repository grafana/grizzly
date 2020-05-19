package dash

import (
	"errors"
	"net/url"
	"os"
)

// Config provides configuration to `grafana-dash`
type Config struct {
	GrafanaURL  string
	JsonnetPath string
}

// ParseEnvironment parses necessary environment variables
func ParseEnvironment() (*Config, error) {
	var config Config
	if grafanaUrl, exists := os.LookupEnv("GRAFANA_URL"); exists {
		u, err := url.Parse(grafanaUrl)
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
	} else {
		return nil, errors.New("Must set GRAFANA_URL environment variable.")
	}
	return &config, nil
}
