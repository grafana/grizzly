package main

import (
	"fmt"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/spf13/viper"
)

func configCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "config <sub-command>",
		Short: "Show, select or configure configuration",
		Args:  cli.ArgsExact(0),
	}
	cmd.AddCommand(configPathCmd())
	cmd.AddCommand(currentContextCmd())
	cmd.AddCommand(useContextCmd())
	cmd.AddCommand(getContextsCmd())
	cmd.AddCommand(configImportCmd())
	cmd.AddCommand(getConfigCmd())
	cmd.AddCommand(setCmd())
	cmd.AddCommand(createContextCmd())
	return cmd
}

func configPathCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "path",
		Short: "Print the path to the configuration file",
		Args:  cli.ArgsExact(0),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		fmt.Println(viper.ConfigFileUsed())
		return nil
	}
	return initialiseLogging(cmd, &opts)
}

func configImportCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "import configuration",
		Short: "Initialise Grizzly configuration file from environment variables",
		Args:  cli.ArgsExact(0),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.Import()
	}
	return initialiseLogging(cmd, &opts)
}
func currentContextCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "current-context",
		Short: "Show current context",
		Args:  cli.ArgsExact(0),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		conf, err := config.CurrentContext()
		if err != nil {
			return err
		}
		fmt.Println(conf.Name)
		return nil
	}
	return initialiseLogging(cmd, &opts)
}

func useContextCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "use-context",
		Short: "Select a context",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		context := args[0]
		return config.UseContext(context)
	}
	return initialiseLogging(cmd, &opts)
}

func getContextsCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get-contexts",
		Short: "list configured contexts",
		Args:  cli.ArgsRange(0, 1),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.GetContexts()
	}
	return initialiseLogging(cmd, &opts)
}

func getConfigCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get [path]",
		Short: "get the whole configuration for a context or a specific attribute of the configuration",
		Args:  cli.ArgsRange(0, 1),
	}
	var opts grizzly.LoggingOpts
	var output string
	cmd.Flags().StringVarP(&output, "output", "o", "yaml", "Output format")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		val, err := config.Get(path, output)
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	}
	return initialiseLogging(cmd, &opts)
}

func setCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "set configuration value",
		Short: "set a configuration value",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.Set(args[0], args[1])
	}
	return initialiseLogging(cmd, &opts)
}

func createContextCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "create-context",
		Short: "Create a configuration context",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.CreateContext(args[0])
	}
	return initialiseLogging(cmd, &opts)
}
