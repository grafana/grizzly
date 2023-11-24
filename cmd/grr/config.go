package main

import (
	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
)

func configInitCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "init configuration",
		Short: "Initialise Grizzly configuration file",
		Args:  cli.ArgsExact(0),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.Init()
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
		return config.CurrentContext()
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
