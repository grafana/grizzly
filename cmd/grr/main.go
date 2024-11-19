package main

import (
	"errors"
	"os"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/internal/logger"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir"
	"github.com/grafana/grizzly/pkg/syntheticmonitoring"
	log "github.com/sirupsen/logrus"
)

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
		Version: config.Version,
	}

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	})

	config.Initialise()
	err := config.Read()
	if err != nil {
		log.Fatalln(err)
	}

	context, err := config.CurrentContext()
	if err != nil {
		log.Fatalln(err)
	}

	log.AddHook(logger.NewSecretsRedactor(context.Secrets()))

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
		captureCmd(registry),
		snapshotCmd(registry),
		providersCmd(registry),
		configCmd(registry),
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
	providers := []grizzly.Provider{
		grafana.NewProvider(&context.Grafana),
		mimir.NewProvider(&context.Mimir),
		syntheticmonitoring.NewProvider(&context.SyntheticMonitoring),
	}

	return grizzly.NewRegistry(providers)
}
