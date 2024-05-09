package main

import (
	"fmt"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/spf13/viper"
)

// LoggingOpts contains logging options (used in all commands)
type LoggingOpts struct {
	LogLevel string
}

// Opts contains options for most Grizzly commands
type Opts struct {
	LoggingOpts
	Directory    bool // Deprecated: now is gathered with os.Stat(<resource-path>)
	JsonnetPaths []string
	Targets      []string
	OutputFormat string
	IsDir        bool // used internally to denote that the resource path argument pointed at a directory

	// Used for supporting resources without envelopes
	OnlySpec     bool
	HasOnlySpec  bool
	FolderUID    string
	ResourceKind string

	// Used for supporting the proxy server
	OpenBrowser bool
	ProxyPort   int
	CanSave     bool
	Watch       bool
}

func configPathCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "path",
		Short: "Print the path to the configuration file",
		Args:  cli.ArgsExact(0),
	}
	var opts LoggingOpts

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
	var opts LoggingOpts

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
	var opts LoggingOpts

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
		Use:   "use-context <context-name>",
		Short: "Select a context",
		Args:  cli.ArgsExact(1),
	}
	var opts LoggingOpts

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
		Args:  cli.ArgsNone(),
	}
	var opts LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		contexts, err := config.GetContexts()
		if err != nil {
			return err
		}
		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}

		for _, context := range contexts {
			if context == currentContext.Name {
				fmt.Printf("* %s\n", context)
			} else {
				fmt.Printf("  %s\n", context)
			}
		}
		return nil
	}
	return initialiseLogging(cmd, &opts)
}

func getConfigCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get [path]",
		Short: "get the whole configuration for a context or a specific attribute of the configuration",
		Args:  cli.ArgsRange(0, 1),
	}
	var opts LoggingOpts
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
		Use:   "set <key> <value>",
		Short: "Set a configuration value for the current context",
		Args:  cli.ArgsExact(2),
	}
	var opts LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.Set(args[0], args[1])
	}
	return initialiseLogging(cmd, &opts)
}

func unsetCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "unset <key>",
		Short: "Unset a configuration value for the current context",
		Args:  cli.ArgsExact(1),
		Run: func(cmd *cli.Command, args []string) error {
			return config.Unset(args[0])
		},
	}

	return initialiseLogging(cmd, &LoggingOpts{})
}

func createContextCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "create-context <context-name>",
		Short: "Create a configuration context",
		Args:  cli.ArgsExact(1),
	}
	var opts LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return config.CreateContext(args[0])
	}
	return initialiseLogging(cmd, &opts)
}
