package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
	log "github.com/sirupsen/logrus"
)

func getCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get <resource-type>.<resource-uid>",
		Short: "retrieve resource",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		return grizzly.Get(uid)
	}
	return initialiseLogging(cmd, &opts)
}

func listCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "list [-r] [<resource-path>]",
		Short: "list resource keys from file",
		Args:  cli.ArgsRange(0, 1),
	}
	var opts grizzly.Opts
	var isRemote bool
	cmd.Flags().BoolVarP(&isRemote, "remote", "r", false, "list remote resources")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if isRemote {
			if len(args) > 0 {
				notifier.Error(nil, "No resource-path required when listing remote resources")
				return nil
			}
			return grizzly.ListRemote(opts)
		}
		if len(args) == 0 {
			notifier.Error(nil, "resource-path required when listing local resources")
			return nil
		}
		resources, err := grizzly.Parse(args[0], opts)
		if err != nil {
			return err
		}

		return grizzly.List(resources)
	}
	return initialiseCmd(cmd, &opts)
}

func pullCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "pull <resource-path>",
		Short: "Pulls remote resources and writes them to local sources",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return grizzly.Pull(args[0], opts)
	}
	return initialiseCmd(cmd, &opts)
}
func showCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "show <resource-path>",
		Short: "show list of resource types and UIDs",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	isResolved := true
	cmd.Flags().BoolVarP(&isResolved, "no-resolve", "r", false, "do not resolve composable resources")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(args[0], opts)
		if err != nil {
			return err
		}
		if !isResolved {
			resources, err = grizzly.Resolve(resources)
			if err != nil {
				return err
			}
		}
		return grizzly.Show(resources)
	}
	return initialiseCmd(cmd, &opts)
}

func diffCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <resource-path>",
		Short: "compare local and remote resources",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(args[0], opts)
		if err != nil {
			return err
		}
		resources, err = grizzly.Resolve(resources)
		if err != nil {
			return err
		}
		return grizzly.Diff(resources)
	}
	return initialiseCmd(cmd, &opts)
}

func applyCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <resource-path>",
		Short: "apply local resources to remote endpoints",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(args[0], opts)
		if err != nil {
			return err
		}
		resources, err = grizzly.Resolve(resources)
		if err != nil {
			return err
		}
		return grizzly.Apply(resources)
	}
	return initialiseCmd(cmd, &opts)
}

type jsonnetWatchParser struct {
	resourcePath string
	opts         grizzly.Opts
}

func (p *jsonnetWatchParser) Name() string {
	return p.resourcePath
}

func (p *jsonnetWatchParser) Parse() (grizzly.Resources, error) {
	resources, err := grizzly.Parse(p.resourcePath, p.opts)
	if err != nil {
		return nil, err
	}
	return grizzly.Resolve(resources)
}

func watchCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <resource-path>",
		Short: "watch dir recursively for file changes and apply selected resource path",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		parser := &jsonnetWatchParser{
			resourcePath: args[1],
			opts:         opts,
		}

		watchDir := args[0]

		return grizzly.Watch(watchDir, parser)
	}
	return initialiseCmd(cmd, &opts)
}

func previewCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "preview <resource-path>",
		Short: "upload a snapshot to preview the rendered file",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	expires := cmd.Flags().IntP("expires", "e", 0, "when the preview should expire. Default 0 (never)")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(args[0], opts)
		if err != nil {
			return err
		}
		resources, err = grizzly.Resolve(resources)
		if err != nil {
			return err
		}

		previewOpts := &grizzly.PreviewOpts{
			ExpiresSeconds: *expires,
		}

		return grizzly.Preview(resources, previewOpts)
	}
	return initialiseCmd(cmd, &opts)
}

func exportCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "export <resource-path> <dashboard-dir>",
		Short: "render resources and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		dashboardDir := args[1]
		resources, err := grizzly.Parse(args[0], opts)
		if err != nil {
			return err
		}
		resources, err = grizzly.Resolve(resources)
		if err != nil {
			return err
		}
		return grizzly.Export(dashboardDir, resources)
	}
	return initialiseCmd(cmd, &opts)
}

func providersCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "providers",
		Short: "Lists all providers registered with Grizzly",
		Args:  cli.ArgsExact(0),
	}
	var opts grizzly.LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		f := "%s\t%s\n"
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(w, f, "API VERSION", "KIND")
		for _, provider := range grizzly.Registry.Providers {
			for _, handler := range provider.GetHandlers() {
				fmt.Fprintf(w, f, provider.APIVersion(), handler.Kind())
			}
		}
		return w.Flush()
	}
	return initialiseLogging(cmd, &opts)
}

func initialiseCmd(cmd *cli.Command, opts *grizzly.Opts) *cli.Command {
	cmd.Flags().BoolVarP(&opts.Directory, "directory", "d", false, "treat resource path as a directory")
	cmd.Flags().StringSliceVarP(&opts.Targets, "target", "t", nil, "resources to target")
	cmd.Flags().StringSliceVarP(&opts.JsonnetPaths, "jpath", "J", getDefaultJsonnetFolders(), "Specify an additional library search dir (right-most wins)")
	return initialiseLogging(cmd, &opts.LoggingOpts)
}

func initialiseLogging(cmd *cli.Command, loggingOpts *grizzly.LoggingOpts) *cli.Command {
	cmd.Flags().StringVarP(&loggingOpts.LogLevel, "log-level", "l", log.InfoLevel.String(), "info, debug, warning, error")
	cmdRun := cmd.Run
	cmd.Run = func(cmd *cli.Command, args []string) error {
		logLevel, err := log.ParseLevel(loggingOpts.LogLevel)
		if err != nil {
			return err
		}
		log.SetLevel(logLevel)
		return cmdRun(cmd, args)
	}

	return cmd
}

func getDefaultJsonnetFolders() []string {
	return []string{"vendor", "lib", "."}
}
