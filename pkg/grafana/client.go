package grafana

import (
	"fmt"
	"net/url"
	"os"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
)

var client *gclient.GrafanaHTTPAPI

func GetClient(conf config.GrafanaConfig) (*gclient.GrafanaHTTPAPI, error) {
	if client != nil {
		return client, nil
	}
	exists, err := config.Exists()
	if err != nil {
		return nil, fmt.Errorf("Error locating configuration file: %v", err)
	}
	if exists {
		parsedUrl, err := url.Parse(conf.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid Grafana URL")
		}

		transportConfig := gclient.DefaultTransportConfig().WithHost(parsedUrl.Host).WithSchemes([]string{parsedUrl.Scheme})
		if conf.Token != "" {
			if conf.User != "" {
				transportConfig.BasicAuth = url.UserPassword(conf.User, conf.Token)
			} else {
				transportConfig.APIKey = conf.Token
			}
		}
		grafanaClient := gclient.NewHTTPClientWithConfig(nil, transportConfig)
		return grafanaClient, nil
	} else {
		grafanaURL, exists := os.LookupEnv("GRAFANA_URL")
		if !exists {
			return nil, fmt.Errorf("Please configure Grizzly using grr config")
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
		notifier.Warn(nil, "Using environment variables for configuration is deprecated. Please use grr config to configure contexts.")
		grafanaClient := gclient.NewHTTPClientWithConfig(nil, transportConfig)
		client = grafanaClient
		return grafanaClient, nil
	}
}
