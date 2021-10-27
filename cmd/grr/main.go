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

func withLoggingOption(cmd *cli.Command) *cli.Command {
	logLevelString := cmd.Flags().StringP("log-level", "l", log.InfoLevel.String(), "info, debug, warning, error")

	cmdRun := cmd.Run
	cmd.Run = func(cmd *cli.Command, args []string) error {
		logLevel, err := log.ParseLevel(*logLevelString)
		if err != nil {
			return err
		}
		log.SetLevel(logLevel)
		return cmdRun(cmd, args)
	}

	return cmd
}

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
		withLoggingOption(getCmd()),
		withLoggingOption(listCmd()),
		withLoggingOption(pullCmd()),
		withLoggingOption(showCmd()),
		withLoggingOption(diffCmd()),
		withLoggingOption(applyCmd()),
		withLoggingOption(watchCmd()),
		withLoggingOption(exportCmd()),
		withLoggingOption(previewCmd()),
		withLoggingOption(providersCmd()),
	)

	// Run!
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
