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
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	terminal "golang.org/x/term"
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
	var format string
	cmd.Flags().BoolVarP(&isRemote, "remote", "r", false, "list remote resources")
	cmd.Flags().StringVarP(&format, "format", "f", "default", "format for listing, one of default, wide, json, yaml")

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

			return grizzly.ListRemote(registry, targets, format)
		}
		if len(args) == 0 {
			notifier.Error(nil, "resource-path required when listing local resources")
			return nil
		}

		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		resources, err := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths).Parse(args[0], grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		})
		if err != nil {
			return err
		}

		return grizzly.List(registry, resources, format)
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
	var continueOnError bool

	cmd.Flags().BoolVarP(&continueOnError, "continue-on-error", "e", false, "don't stop pulling on error")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		eventsRecorder := getEventsRecorder(opts)
		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}

		targets := currentContext.GetTargets(opts.Targets)

		err = grizzly.Pull(registry, args[0], onlySpec, format, targets, continueOnError, eventsRecorder)

		notifier.Info(nil, eventsRecorder.Summary().AsString("resource"))

		// errors are already displayed by the `eventsRecorder`, so we return a
		// "silent" one to ensure that the exit code will be non-zero
		if err != nil {
			return silentError{Err: err}
		}

		return nil
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

		resources, err := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths).Parse(args[0], grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		})
		if err != nil {
			return err
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

		resources, err := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths).Parse(args[0], grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		})
		if err != nil {
			return err
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
	var continueOnError bool

	cmd.Flags().BoolVarP(&continueOnError, "continue-on-error", "e", false, "don't stop apply on first error")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		eventsRecorder := getEventsRecorder(opts)
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}

		targets := currentContext.GetTargets(opts.Targets)
		parser := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths, grizzly.ParserContinueOnError(continueOnError))

		resources, parseErr := parser.Parse(args[0], grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		})

		if parseErr != nil {
			var parseErrors []error
			if merr, ok := parseErr.(*multierror.Error); ok {
				parseErrors = merr.Errors
			} else {
				parseErrors = []error{parseErr}
			}

			for _, e := range parseErrors {
				notifier.Error(nil, e.Error())
			}
		}

		if parseErr != nil && !continueOnError {
			return silentError{Err: parseErr}
		}

		notifier.Info(nil, fmt.Sprintf("Applying %s", grizzly.Pluraliser(resources.Len(), "resource")))

		applyErr := grizzly.Apply(registry, resources, continueOnError, eventsRecorder)

		notifier.Info(nil, eventsRecorder.Summary().AsString("resource"))

		// errors are already displayed by the `eventsRecorder`, so we return a
		// "silent" one to ensure that the exit code will be non-zero
		if parseErr != nil || applyErr != nil {
			return silentError{Err: errors.Join(parseErr, applyErr)}
		}

		return nil
	}

	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
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

		watchDir, resourcePath := args[0], args[1]

		trailRecorder := grizzly.NewWriterRecorder(os.Stdout, grizzly.EventToPlainText)

		parser := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths, grizzly.ParserContinueOnError(true))
		parserOpts := grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		}
		return grizzly.Watch(registry, watchDir, resourcePath, parser, parserOpts, trailRecorder)
	}
	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

func snapshotCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "snapshot <resource-path>",
		Short: "upload a snapshot to preview resources",
		Args:  cli.ArgsExact(1),
	}
	var opts Opts
	expires := cmd.Flags().IntP("expires", "e", 0, "when the snapshot should expire. Default 0 (never)")

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
		parser := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths, grizzly.ParserContinueOnError(false))

		resources, parseErr := parser.Parse(args[0], grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		})

		if parseErr != nil {
			var parseErrors []error
			if merr, ok := parseErr.(*multierror.Error); ok {
				parseErrors = merr.Errors
			} else {
				parseErrors = []error{parseErr}
			}

			for _, e := range parseErrors {
				notifier.Error(nil, e.Error())
			}
			return silentError{Err: parseErr}
		}
		return grizzly.Snapshot(registry, resources, *expires)
	}
	return initialiseCmd(cmd, &opts)
}

func serveCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "serve <resources>",
		Short: "Run Grizzly server",
		Args:  cli.ArgsAny(),
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

		resourcesPath := ""
		watchPaths := args

		if len(args) > 0 {
			resourcesPath = args[0]
		}

		if opts.WatchScript != "" {
			resourcesPath = ""
		}
		if len(args) > 1 {
			watchPaths = args[1:]
		}

		targets := currentContext.GetTargets(opts.Targets)
		parser := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths, grizzly.ParserContinueOnError(true))
		parserOpts := grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		}

		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}

		server, err := grizzly.NewGrizzlyServer(registry, resourcesPath, opts.ProxyPort)
		if err != nil {
			return err
		}
		server.SetParser(parser, parserOpts)
		server.SetContext(currentContext.Name)
		server.SetFormatting(onlySpec, format)
		if opts.Watch {
			server.Watch(watchPaths, opts.JsonnetPaths)
			if opts.WatchScript != "" {
				server.WatchScript(opts.WatchScript)
			}
		}
		if opts.OpenBrowser {
			server.OpenBrowser()
		}
		return server.Start()
	}
	cmd.Flags().BoolVarP(&opts.Watch, "watch", "w", false, "Watch filesystem for changes")
	cmd.Flags().BoolVarP(&opts.OpenBrowser, "open-browser", "b", false, "Open Grizzly in default browser")
	cmd.Flags().IntVarP(&opts.ProxyPort, "port", "p", 8080, "Port on which the server will listen")
	cmd.Flags().StringVarP(&opts.WatchScript, "script", "S", "", "Script to execute on filesystem change")
	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}

func exportCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "export <resource-path> <export-dir>",
		Short: "render resources and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	var opts Opts
	var continueOnError bool

	cmd.Flags().BoolVarP(&continueOnError, "continue-on-error", "e", false, "don't stop exporting on error")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resourcePath := args[0]
		exportDir := args[1]
		resourceKind, folderUID, err := getOnlySpec(opts)
		if err != nil {
			return err
		}

		currentContext, err := config.CurrentContext()
		if err != nil {
			return err
		}

		targets := currentContext.GetTargets(opts.Targets)

		resources, err := grizzly.DefaultParser(registry, targets, opts.JsonnetPaths, grizzly.ParserContinueOnError(continueOnError)).Parse(resourcePath, grizzly.ParserOptions{
			DefaultResourceKind: resourceKind,
			DefaultFolderUID:    folderUID,
		})
		if err != nil {
			return err
		}

		format, onlySpec, err := getOutputFormat(opts)
		if err != nil {
			return err
		}

		eventsRecorder := getEventsRecorder(opts)

		err = grizzly.Export(eventsRecorder, registry, exportDir, resources, onlySpec, format, continueOnError)

		notifier.Info(nil, eventsRecorder.Summary().AsString("resource"))

		// errors are already displayed by the `eventsRecorder`, so we return a
		// "silent" one to ensure that the exit code will be non-zero
		if err != nil {
			return silentError{Err: err}
		}

		return nil
	}
	cmd = initialiseOnlySpec(cmd, &opts)
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

func configCmd(registry grizzly.Registry) *cli.Command {
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
	cmd.AddCommand(unsetCmd())
	cmd.AddCommand(createContextCmd())
	cmd.AddCommand(checkCmd(registry))
	return cmd
}

func initialiseCmd(cmd *cli.Command, opts *Opts) *cli.Command {
	// Keep the old flags for backwards compatibility
	cmd.Flags().BoolVarP(&opts.Directory, "directory", "d", false, "treat resource path as a directory")
	if err := cmd.Flags().MarkDeprecated("directory", "now it is inferred from the operating system"); err != nil {
		log.Fatal(err)
	}

	cmd.Flags().StringSliceVarP(&opts.Targets, "target", "t", nil, "resources to target")
	cmd.Flags().StringSliceVarP(&opts.JsonnetPaths, "jpath", "J", getDefaultJsonnetFolders(), "Specify an additional library search dir (right-most wins)")
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "", "Output format")

	cmd.Flags().BoolVar(&opts.DisableStats, "disable-reporting", false, "disable sending of anonymous usage stats to Grafana Labs")

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

func getEventsRecorder(opts Opts) grizzly.EventsRecorder {
	wr := grizzly.NewWriterRecorder(os.Stdout, getEventFormatter())
	if opts.DisableStats || config.UsageStatsDisabled() {
		return wr
	}
	return grizzly.NewUsageRecorder(wr)
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

func getEventFormatter() grizzly.EventFormatter {
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		return grizzly.EventToColoredText
	}

	return grizzly.EventToPlainText
}
