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
	providerInitFuncs := map[string]func() (grizzly.Provider, error){
		"Grafana":              func() (grizzly.Provider, error) { return grafana.NewProvider(&context.Grafana) },
		"Mimir":                func() (grizzly.Provider, error) { return mimir.NewProvider(&context.Mimir) },
		"Synthetic Monitoring": func() (grizzly.Provider, error) { return syntheticmonitoring.NewProvider(&context.SyntheticMonitoring) },
	}

	providers := []grizzly.Provider{}
	initMessage := "Providers:"
	for name, initFunc := range providerInitFuncs {
		provider, err := initFunc()
		if err != nil {
			initMessage += "\n  " + name + " - inactive: " + err.Error()
			continue
		}
		initMessage += "\n  " + name + " - active"
		providers = append(providers, provider)
	}
	log.Info(initMessage)
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
