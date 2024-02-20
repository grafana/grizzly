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

const generalFolderUID = "general"

func getCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "get <resource-type>.<resource-uid>",
		Short: "retrieve resource",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		return grizzly.Get(registry, uid, opts)
	}
	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

func listCmd(registry grizzly.Registry) *cli.Command {
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
			return grizzly.ListRemote(registry, opts)
		}
		if len(args) == 0 {
			notifier.Error(nil, "resource-path required when listing local resources")
			return nil
		}
		resources, err := grizzly.Parse(registry, args[0], &opts)
		if err != nil {
			return err
		}

		return grizzly.List(registry, resources)
	}
	return initialiseCmd(cmd, &opts)
}

func pullCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "pull <resource-path>",
		Short: "Pulls remote resources and writes them to local sources",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		return grizzly.Pull(registry, args[0], opts)
	}

	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

func showCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "show <resource-path>",
		Short: "show list of resource types and UIDs",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], &opts)
		if err != nil {
			return err
		}
		return grizzly.Show(registry, resources, opts)
	}
	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

func diffCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <resource-path>",
		Short: "compare local and remote resources",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], &opts)
		if err != nil {
			return err
		}
		return grizzly.Diff(registry, resources, opts)
	}
	return initialiseCmd(cmd, &opts)
}

func applyCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:     "apply <resource-path>",
		Aliases: []string{"push"},
		Short:   "apply local resources to remote endpoints",
		Args:    cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], &opts)
		if err != nil {
			return err
		}
		return grizzly.Apply(registry, resources)
	}

	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

type jsonnetWatchParser struct {
	resourcePath string
	registry     grizzly.Registry
	opts         grizzly.Opts
}

func (p *jsonnetWatchParser) Name() string {
	return p.resourcePath
}

func (p *jsonnetWatchParser) Parse() (grizzly.Resources, error) {
	return grizzly.Parse(p.registry, p.resourcePath, &p.opts)
}

func watchCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <resource-path>",
		Short: "watch dir recursively for file changes and apply selected resource path",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		parser := &jsonnetWatchParser{
			resourcePath: args[1],
			registry:     registry,
			opts:         opts,
		}

		watchDir := args[0]

		return grizzly.Watch(registry, watchDir, parser)
	}
	return initialiseCmd(cmd, &opts)
}

func previewCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "preview <resource-path>",
		Short: "upload a snapshot to preview the rendered file",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	expires := cmd.Flags().IntP("expires", "e", 0, "when the preview should expire. Default 0 (never)")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], &opts)
		if err != nil {
			return err
		}

		previewOpts := &grizzly.PreviewOpts{
			ExpiresSeconds: *expires,
		}

		return grizzly.Preview(registry, resources, previewOpts)
	}
	return initialiseCmd(cmd, &opts)
}

func serveCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "serve <resources>",
		Short: "Run Grizzly server",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		parser := &jsonnetWatchParser{
			resourcePath: args[0],
			registry:     registry,
			opts:         opts,
		}
		return grizzly.Serve(registry, parser, args[0], opts)
	}
	cmd.Flags().BoolVarP(&opts.OpenBrowser, "open-browser", "b", false, "Open Grizzly in default browser")
	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

func exportCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "export <resource-path> <dashboard-dir>",
		Short: "render resources and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		dashboardDir := args[1]
		resources, err := grizzly.Parse(registry, args[0], &opts)
		if err != nil {
			return err
		}
		return grizzly.Export(registry, dashboardDir, resources, opts)
	}
	return initialiseCmd(cmd, &opts)
}

func providersCmd(registry grizzly.Registry) *cli.Command {
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
		for _, provider := range registry.Providers {
			for _, handler := range provider.GetHandlers() {
				fmt.Fprintf(w, f, provider.APIVersion(), handler.Kind())
			}
		}
		return w.Flush()
	}

	return initialiseLogging(cmd, &opts)
}

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

func initialiseCmd(cmd *cli.Command, opts *grizzly.Opts) *cli.Command {
	// Keep the old flags for backwards compatibility
	cmd.Flags().BoolVarP(&opts.Directory, "directory", "d", false, "treat resource path as a directory")
	cmd.Flags().MarkDeprecated("directory", "now it is inferred from the operating system")

	cmd.Flags().StringSliceVarP(&opts.Targets, "target", "t", nil, "resources to target")
	cmd.Flags().StringSliceVarP(&opts.JsonnetPaths, "jpath", "J", getDefaultJsonnetFolders(), "Specify an additional library search dir (right-most wins)")
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "", "Output format")

	return initialiseLogging(cmd, &opts.LoggingOpts)
}

func initialiseOnlySpec(cmd *cli.Command, opts *grizzly.Opts) *cli.Command {
	cmd.Flags().BoolVarP(&opts.OnlySpec, "only-spec", "s", false, "this flag is only used for dashboards to output the spec")
	cmd.Flags().StringVarP(&opts.FolderUID, "folder", "f", generalFolderUID, "folder to push dashboards to")
	cmd.Flags().StringVarP(&opts.ResourceKind, "kind", "k", "", "Kind to use for resources. Required by --only-spec")

	cmdRun := cmd.Run
	cmd.Run = func(cmd *cli.Command, args []string) error {
		opts.HasOnlySpec = cmd.Flags().Changed("only-spec")
		return cmdRun(cmd, args)
	}

	return cmd
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
