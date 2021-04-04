package main

import (
	"log"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Version is the current version of the grr command.
// To be overwritten at build time
var Version = "dev"

func main() {
	log.SetFlags(log.Ltime)

	rootCmd := &cli.Command{
		Use:     "grr",
		Short:   "Grizzly",
		Version: Version,
	}

	registry, err := GetProviderRegistry()
	if err != nil {
		log.Fatalln(err)
	}

	// workflow commands
	rootCmd.AddCommand(
		getCmd(registry),
		listCmd(registry),
		pullCmd(registry),
		showCmd(registry),
		diffCmd(registry),
		applyCmd(registry),
		watchCmd(registry),
		listenCmd(registry),
		exportCmd(registry),
		previewCmd(registry),
		providersCmd(registry),
	)

	// Run!
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

// GetProviderRegistry registers all known providers
func GetProviderRegistry() (grizzly.Registry, error) {
	registry := grizzly.NewProviderRegistry()
	registry.RegisterProvider(&grafana.Provider{})
	return registry, nil
}
