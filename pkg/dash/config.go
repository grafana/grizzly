package dash

import (
	"errors"
	"fmt"
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
	if url, exists := os.LookupEnv("GRAFANA_URL"); exists {
		config.GrafanaURL = url
	} else {
		protocol, protocolExists := os.LookupEnv("GRAFANA_PROTOCOL")
		user, userExists := os.LookupEnv("GRAFANA_USER")
		token, tokenExists := os.LookupEnv("GRAFANA_TOKEN")
		host, hostExists := os.LookupEnv("GRAFANA_HOST")
		path, pathExists := os.LookupEnv("GRAFANA_PATH")
		if !hostExists {
			return nil, errors.New("Either GRAFANA_URL or GRAFANA_HOST required")
		}
		if !protocolExists {
			protocol = "https"
		}
		auth := ""
		if userExists && tokenExists {
			auth = fmt.Sprintf("%s:%s", user, token)
		}
		if pathExists {
			path = "/" + path
		}

		config.GrafanaURL = fmt.Sprintf("%s://%s@%s%s", protocol, auth, host, path)
	}
	return &config, nil
}
