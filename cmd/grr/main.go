package main

import (
	"fmt"
	"log"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Version is the current version of the grr command.
// To be overwritten at build time
var Version = "dev"

func main() {
	log.SetFlags(0)

	rootCmd := &cli.Command{
		Use:     "grr",
		Short:   "Grizzly",
		Version: Version,
	}

	registry, err := GetProviderRegistry()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Provider count", len(registry.GetProviders()))

	for i, provider := range registry.GetProviders() {
		path := provider.GetJSONPath()
		fmt.Printf("%02d: %s/%s\n", i+1, provider.GetName(), path)
	}

	config := grizzly.Config{
		Registry: registry,
	}
	// workflow commands
	rootCmd.AddCommand(
		getCmd(config),
		listCmd(config),
		showCmd(config),
		diffCmd(config),
		applyCmd(config),
		watchCmd(config),
		exportCmd(config),
		previewCmd(config),
	)

	// Run!
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

// GetProviderRegistry registers all known providers
func GetProviderRegistry() (grizzly.Registry, error) {
	registry := grizzly.NewProviderRegistry()
	registry.RegisterProvider(grafana.NewDashboardProvider())
	registry.RegisterProvider(grafana.NewDatasourceProvider())
	//registry.RegisterProvider(grafana.NewPluginProvider())
	//registry.RegisterProvider(grafana.NewMixinProvider())
	return registry, nil
}
