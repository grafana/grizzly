package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/config"
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
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}
		return grizzly.Get(registry, uid, onlySpec, format)
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
	var opts Opts
	var isRemote bool
	cmd.Flags().BoolVarP(&isRemote, "remote", "r", false, "list remote resources")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)

		if isRemote {
			if len(args) > 0 {
				notifier.Error(nil, "No resource-path required when listing remote resources")
				return nil
			}

			return grizzly.ListRemote(registry, targets)
		}
		if len(args) == 0 {
			notifier.Error(nil, "resource-path required when listing local resources")
			return nil
		}

		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		resources, errs := grizzly.Parse(registry, args[0], resourceKind, folderUID, targets, opts.JsonnetPaths)
		if errs != nil {
			return errors.Join(errs...)
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
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)
		return grizzly.Pull(registry, args[0], onlySpec, format, targets)
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
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)

		resources, errs := grizzly.Parse(registry, args[0], resourceKind, folderUID, targets, opts.JsonnetPaths)
		if errs != nil {
			return errors.Join(errs...)
		}

		format, _, err := getOutputFormat(opts)
		if err != nil {
			return err
		}
		return grizzly.Show(registry, resources, format)
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
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)

		resources, errs := grizzly.Parse(registry, args[0], resourceKind, folderUID, targets, opts.JsonnetPaths)
		if errs != nil {
			return errors.Join(errs...)
		}
		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}

		return grizzly.Diff(registry, resources, onlySpec, format)
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
	var opts Opts
	continueOnError := false
	cmd.Flags().BoolVarP(&continueOnError, "continue-on-error", "e", false, "don't stop apply on first error")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)

		resources, errs := grizzly.Parse(registry, args[0], resourceKind, folderUID, targets, opts.JsonnetPaths)
		if errs != nil {
			return errors.Join(errs...)
		}
		return grizzly.Apply(registry, resources, continueOnError)
	}

	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

type jsonnetWatchParser struct {
	resourcePath string
	registry     grizzly.Registry
	resourceKind string
	folderUID    string
	targets      []string
	jsonnetPaths []string
}

func (p *jsonnetWatchParser) Name() string {
	return p.resourcePath
}

func (p *jsonnetWatchParser) Parse() (grizzly.Resources, []error) {
	return grizzly.Parse(p.registry, p.resourcePath, p.resourceKind, p.folderUID, p.targets, p.jsonnetPaths)
}

func watchCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <resource-path>",
		Short: "watch dir recursively for file changes and apply selected resource path",
		Args:  cli.ArgsExact(2),
	}
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)
		parser := &jsonnetWatchParser{
			resourcePath: args[1],
			registry:     registry,
			resourceKind: resourceKind,
			folderUID:    folderUID,
			targets:      targets,
			jsonnetPaths: opts.JsonnetPaths,
		}

		watchDir := args[0]

		return grizzly.Watch(registry, watchDir, parser)
	}
	return initialiseCmd(cmd, &opts)
}

func serveCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "serve <resources>",
		Short: "Run Grizzly server",
		Args:  cli.ArgsExact(1),
	}
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)
		parser := &jsonnetWatchParser{
			resourcePath: args[0],
			registry:     registry,
			resourceKind: resourceKind,
			folderUID:    folderUID,
			targets:      targets,
			jsonnetPaths: opts.JsonnetPaths,
		}
		port := 8080
		if opts.ProxyPort != 0 {
			port = opts.ProxyPort
		}
		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}
		return grizzly.Serve(registry, parser, args[0], port, opts.OpenBrowser, onlySpec, format)
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
	var opts Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourcePath := args[0]
		dashboardDir := args[1]
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}
		targets := currentContext.GetTargets(opts.Targets)

		resources, errs := grizzly.Parse(registry, resourcePath, resourceKind, folderUID, targets, opts.JsonnetPaths)
		if errs != nil {
			return errors.Join(errs...)
		}
		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}
		return grizzly.Export(registry, dashboardDir, resources, onlySpec, format)
	}
	return initialiseCmd(cmd, &opts)
}

func providersCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "providers",
		Short: "Lists all providers registered with Grizzly",
		Args:  cli.ArgsExact(0),
	}
	var opts LoggingOpts

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

func initialiseCmd(cmd *cli.Command, opts *Opts) *cli.Command {
	// Keep the old flags for backwards compatibility
	cmd.Flags().BoolVarP(&opts.Directory, "directory", "d", false, "treat resource path as a directory")
	cmd.Flags().MarkDeprecated("directory", "now it is inferred from the operating system")

	cmd.Flags().StringSliceVarP(&opts.Targets, "target", "t", nil, "resources to target")
	cmd.Flags().StringSliceVarP(&opts.JsonnetPaths, "jpath", "J", getDefaultJsonnetFolders(), "Specify an additional library search dir (right-most wins)")
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "", "Output format")

	return initialiseLogging(cmd, &opts.LoggingOpts)
}

func initialiseOnlySpec(cmd *cli.Command, opts *Opts) *cli.Command {
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

func initialiseLogging(cmd *cli.Command, loggingOpts *LoggingOpts) *cli.Command {
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
func getOutputFormat(opts Opts) (string, bool, error) {
	var onlySpec bool
	context, err := config.CurrentContext()
	if err != nil {
		return "", false, err
	}
	if opts.HasOnlySpec {
		onlySpec = opts.OnlySpec
	} else {
		onlySpec = context.OnlySpec
	}
	if opts.OutputFormat != "" {
		return opts.OutputFormat, onlySpec, nil
	} else if context.OutputFormat != "" {
		return context.OutputFormat, onlySpec, nil
	}
	return "yaml", onlySpec, nil
}

func getOnlySpec(opts Opts) (string, string, error) {
	context, err := config.CurrentContext()
	if err != nil {
		return "", "", err
	}
	kind := context.ResourceKind
	if kind == "" {
		kind = opts.ResourceKind
	}
	folderUID := context.FolderUID
	if folderUID == "" {
		folderUID = opts.FolderUID
	}
	return kind, folderUID, nil
}
