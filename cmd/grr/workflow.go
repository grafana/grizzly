package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/spf13/pflag"
)

func getCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "get <resource-type>.<resource-uid>",
		Short: "retrieve resource",
		Args:  cli.ArgsExact(1),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		return grizzly.Get(registry, uid)
	}
	return cmd
}

func listCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "list [-r] [<resource-path>]",
		Short: "list resource keys from file",
		Args:  cli.ArgsRange(0, 1),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	var isRemote bool
	cmd.Flags().BoolVarP(&isRemote, "remote", "r", false, "list remote resources")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if isRemote {
			if len(args) > 0 {
				registry.Notifier().Error(nil, "No resource-path required when listing remote resources")
				return nil
			}
			return grizzly.ListRemote(registry, opts)
		}
		if len(args) == 0 {
			registry.Notifier().Error(nil, "resource-path required when listing local resources")
			return nil
		}
		resources, err := grizzly.Parse(registry, args[0], opts)
		if err != nil {
			return err
		}

		return grizzly.List(registry, resources)
	}
	return cmd
}

func pullCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "pull <resource-path>",
		Short: "Pulls remote resources and writes them to local sources",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	cmd.Run = func(cmd *cli.Command, args []string) error {
		return grizzly.Pull(registry, args[0], opts)
	}
	return cmd
}
func showCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "show <resource-path>",
		Short: "show list of resource types and UIDs",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], opts)
		if err != nil {
			return err
		}
		return grizzly.Show(registry, resources)
	}
	return cmd
}

func diffCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <resource-path>",
		Short: "compare local and remote resources",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], opts)
		if err != nil {
			return err
		}
		return grizzly.Diff(registry, resources)
	}
	return cmd
}

func applyCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <resource-path>",
		Short: "apply local resources to remote endpoints",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], opts)
		if err != nil {
			return err
		}
		return grizzly.Apply(registry, resources)
	}
	return cmd
}

type jsonnetWatchParser struct {
	resourcePath string
	opts         grizzly.Opts
}

func (p *jsonnetWatchParser) Name() string {
	return p.resourcePath
}

func (p *jsonnetWatchParser) Parse(registry grizzly.Registry) (grizzly.Resources, error) {
	return grizzly.Parse(registry, p.resourcePath, p.opts)
}

func watchCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <resource-path>",
		Short: "watch for file changes and apply",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	cmd.Run = func(cmd *cli.Command, args []string) error {
		parser := &jsonnetWatchParser{
			resourcePath: args[1],
			opts:         opts,
		}
		watchDir := args[0]

		return grizzly.Watch(registry, watchDir, parser)
	}
	return cmd
}

func listenCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "listen <uid-to-watch> <output-file>",
		Short: "listen for file changes on remote and save locally",
		Args:  cli.ArgsExact(2),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		filename := args[1]
		return grizzly.Listen(registry, uid, filename)
	}
	return cmd
}

func previewCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "preview <resource-path>",
		Short: "upload a snapshot to preview the rendered file",
		Args:  cli.ArgsExact(1),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	expires := cmd.Flags().IntP("expires", "e", 0, "when the preview should expire. Default 0 (never)")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		resources, err := grizzly.Parse(registry, args[0], opts)
		if err != nil {
			return err
		}

		previewOpts := &grizzly.PreviewOpts{
			ExpiresSeconds: *expires,
		}

		return grizzly.Preview(registry, resources, previewOpts)
	}
	return cmd
}

func exportCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "export <resource-path> <dashboard-dir>",
		Short: "render resources and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts
	defaultGrizzlyFlags(&opts, cmd.Flags())
	cmd.Run = func(cmd *cli.Command, args []string) error {
		dashboardDir := args[1]
		resources, err := grizzly.Parse(registry, args[0], opts)
		if err != nil {
			return err
		}
		return grizzly.Export(registry, dashboardDir, resources)
	}
	return cmd
}

func providersCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "providers",
		Short: "Lists all providers registered with Grizzly",
		Args:  cli.ArgsExact(0),
	}
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
	return cmd
}

func defaultGrizzlyFlags(opts *grizzly.Opts, fs *pflag.FlagSet) {
	fs.BoolVarP(&opts.Directory, "directory", "d", false, "treat resource path as a directory")
	fs.StringSliceVarP(&opts.Targets, "target", "t", nil, "resources to target")
	fs.StringSliceVarP(&opts.JsonnetPaths, "jpath", "J", getDefaultJsonnetFolders(), "Specify an additional library search dir (right-most wins)")
}

func getDefaultJsonnetFolders() []string {
	return []string{"vendor", "lib", "."}
}
