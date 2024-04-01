package main

import (
	"errors"
	"os"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir"
	"github.com/grafana/grizzly/pkg/syntheticmonitoring"
	log "github.com/sirupsen/logrus"
)

// Version is the current version of the grr command.
// To be overwritten at build time
var Version = "dev"

type silentError struct {
	Err error
}

func (err silentError) Is(target error) bool {
	var e silentError
	return errors.As(target, &e)
}

func (err silentError) Error() string {
	return err.Err.Error()
}

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

	registry := createRegistry(context)
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
		snapshotCmd(registry),
		providersCmd(registry),
		configCmd(),
		serveCmd(registry),
		selfUpdateCmd(),
	)

	// Run!
	if err = rootCmd.Execute(); err != nil {
		if errors.Is(err, silentError{}) {
			log.Debugf("Silent error: %s", err)
			os.Exit(1)
		} else {
			log.Fatalln(err)
		}
	}
}

func createRegistry(context *config.Context) grizzly.Registry {
	providerInitFuncs := []func() (grizzly.Provider, error){
		func() (grizzly.Provider, error) { return grafana.NewProvider(&context.Grafana) },
		func() (grizzly.Provider, error) { return mimir.NewProvider(&context.Mimir) },
		func() (grizzly.Provider, error) { return syntheticmonitoring.NewProvider(&context.SyntheticMonitoring) },
	}

	var providers []grizzly.Provider
	initMessage := "Providers:"
	for _, initFunc := range providerInitFuncs {
		provider, err := initFunc()
		if err != nil {
			initMessage += "\n  " + provider.Name() + " - inactive: " + err.Error()
			continue
		}
		initMessage += "\n  " + provider.Name() + " - active"
		providers = append(providers, provider)
	}
	log.Info(initMessage)
	return grizzly.NewRegistry(providers)
}
