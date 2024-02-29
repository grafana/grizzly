package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir"
	"github.com/grafana/grizzly/pkg/syntheticmonitoring"
)

// Version is the current version of the grr command.
// To be overwritten at build time
var Version = "dev"

func main() {

	rootCmd := &cli.Command{
		Use:     "grr",
		Short:   "Grizzly",
		Version: Version,
	}

	config.Initialise()
	err := config.Read()
	if err != nil {
		log.Fatalln(err)
	}

	context, err := config.CurrentContext()
	if err != nil {
		log.Fatalln(err)
	}
	providers := []grizzly.Provider{}

	grafanaProvider, err := grafana.NewProvider(&context.Grafana)
	if err != nil {
		log.Warnf("Grafana provider is not configured: %v", err)
	} else {
		providers = append(providers, grafanaProvider)
	}

	mimirProvider, err := mimir.NewProvider(&context.Mimir)
	if err != nil {
		log.Warnf("Mimir provider is not configured: %v", err)
	} else {
		providers = append(providers, mimirProvider)
	}

	syntheticMonitoringProvider, err := syntheticmonitoring.NewProvider(&context.SyntheticMonitoring)
	if err != nil {
		log.Warnf("Synthetic Monitoring provider is not configured: %v", err)
	} else {
		providers = append(providers, syntheticMonitoringProvider)
	}

	registry := grizzly.NewRegistry(providers)

	// workflow commands
	rootCmd.AddCommand(
		getCmd(registry),
		listCmd(registry),
		pullCmd(registry),
		showCmd(registry),
		diffCmd(registry),
		applyCmd(registry),
		watchCmd(registry),
		exportCmd(registry),
		providersCmd(registry),
		configCmd(),
		serveCmd(registry),
	)

	// Run!
	if err = rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
