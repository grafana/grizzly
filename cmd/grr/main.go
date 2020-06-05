package main

import (
	"log"

	"github.com/go-clix/cli"
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

	// workflow commands
	rootCmd.AddCommand(
		getCmd(),
		listCmd(),
		showCmd(),
		diffCmd(),
		applyCmd(),
		watchCmd(),
		exportCmd(),
	)

	// Run!
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
