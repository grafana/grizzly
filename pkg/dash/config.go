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
	if gu, exists := os.LookupEnv("GRAFANA_URL"); exists {
		config.GrafanaURL = gu
		if token, exists := os.LookupEnv("GRAFANA_TOKEN"); exists {
			u, err := url.Parse(config.GrafanaURL)
			if err != nil {
				return nil, err
			}
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
