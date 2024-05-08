package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
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
	providerInitFuncs := []func() grizzly.Provider{
		func() grizzly.Provider { return grafana.NewProvider(&context.Grafana) },
		func() grizzly.Provider { return mimir.NewProvider(&context.Mimir) },
		func() grizzly.Provider { return syntheticmonitoring.NewProvider(&context.SyntheticMonitoring) },
	}

	var providers []grizzly.Provider

	var providerList []string
	for _, initFunc := range providerInitFuncs {
		provider := initFunc()
		err := provider.Validate()
		if err != nil {
			providerList = append(providerList, fmt.Sprintf("%s - inactive (%s)", provider.Name(), err.Error()))
		} else {
			providerList = append(providerList, provider.Name()+" - active")
		}
		providers = append(providers, provider)
	}
	notifier.InfoStderr(nil, "Providers: "+strings.Join(providerList, ", "))
	return grizzly.NewRegistry(providers)
}
