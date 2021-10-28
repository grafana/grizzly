package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
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

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&grafana.Provider{},
		})

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
