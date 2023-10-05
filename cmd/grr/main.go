package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/go-clix/cli"
	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	log "github.com/sirupsen/logrus"
)

// Version is the current version of the grr command.
// To be overwritten at build time
var Version = "dev"

func main() {
	gclient, err := initGrafanaClient()
	if err != nil {
		log.Fatalln(err)
	}

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			grafana.NewProvider(gclient),
		})

	rootCmd := &cli.Command{
		Use:     "grr",
		Short:   "Grizzly",
		Version: Version,
	}

	// workflow commands
	rootCmd.AddCommand(
		getCmd(),
		listCmd(),
		pullCmd(),
		showCmd(),
		diffCmd(),
		applyCmd(),
		watchCmd(),
		exportCmd(),
		previewCmd(),
		providersCmd(),
	)

	// Run!
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func initGrafanaClient() (*gapi.Client, error) {
	// Read Grafana url from env
	baseURL, err := getGrafanaBaseURL()
	if err != nil {
		return nil, err
	}

	// Init auth configuration from env
	cfg := initGrafanaAuthConfig()

	// Use HTTP client with timeouts
	cfg.Client, err = grafana.NewHttpClient()
	if err != nil {
		return nil, err
	}

	return gapi.New(baseURL, cfg)
}

func initGrafanaAuthConfig() gapi.Config {
	token, exists := os.LookupEnv("GRAFANA_TOKEN")
	if !exists {
		return gapi.Config{}
	}

	user, exists := os.LookupEnv("GRAFANA_USER")
	if !exists {
		return gapi.Config{APIKey: token}
	}

	return gapi.Config{BasicAuth: url.UserPassword(user, token)}
}

func getGrafanaBaseURL() (string, error) {
	if grafanaURL, exists := os.LookupEnv("GRAFANA_URL"); exists {
		u, err := url.Parse(grafanaURL)
		if err != nil {
			return "", err
		}

		return u.String(), nil
	}

	return "", fmt.Errorf("Require GRAFANA_URL (optionally GRAFANA_TOKEN & GRAFANA_USER")
}
