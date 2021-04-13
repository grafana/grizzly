package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grizzly"
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
		Use:   "list [-d <directory> [<jsonnet-file>]",
		Short: "list resource keys from file",
		Args:  cli.ArgsRange(0, 1),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) > 0 {
			opts.ResourceFile = &args[0]
		}
		resources, err := grizzly.Parse(registry, opts)
		if err != nil {
			return err
		}

		return grizzly.List(registry, resources)
	}
	return cmd
}

func pullCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "pull",
		Short: "Pulls remote resources and writes them to local sources",
		Args:  cli.ArgsNone(),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		return grizzly.Pull(registry, opts)
	}
	return cmd
}
func showCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "show [-d <directory>] [<jsonnet-file>]",
		Short: "render Jsonnet as json",
		Args:  cli.ArgsRange(0, 1),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) > 0 {
			opts.ResourceFile = &args[0]
		}
		resources, err := grizzly.Parse(registry, opts)
		if err != nil {
			return err
		}
		return grizzly.Show(registry, resources)
	}
	return cmd
}

func diffCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <jsonnet-file>",
		Short: "compare Jsonnet resources with endpoint(s)",
		Args:  cli.ArgsRange(0, 1),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) > 0 {
			opts.ResourceFile = &args[0]
		}
		resources, err := grizzly.Parse(registry, opts)
		if err != nil {
			return err
		}
		return grizzly.Diff(registry, resources)
	}
	return cmd
}

func applyCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <jsonnet-file>",
		Short: "render Jsonnet and push dashboard(s) to Grafana",
		Args:  cli.ArgsRange(0, 1),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) > 0 {
			opts.ResourceFile = &args[0]
		}
		resources, err := grizzly.Parse(registry, opts)
		if err != nil {
			return err
		}
		return grizzly.Apply(registry, resources)
	}
	return cmd
}

type jsonnetWatchParser struct {
	jsonnetFile string
	opts        grizzly.GrizzlyOpts
}

func (p *jsonnetWatchParser) Name() string {
	return p.jsonnetFile
}

func (p *jsonnetWatchParser) Parse(registry grizzly.Registry) (grizzly.Resources, error) {
	return grizzly.Parse(registry, p.opts)
}

func watchCmd(registry grizzly.Registry) *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <jsonnet-file>",
		Short: "watch for file changes and apply",
		Args:  cli.ArgsExact(2),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		parser := &jsonnetWatchParser{
			jsonnetFile: args[1],
			opts:        opts,
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
		Use:   "preview <jsonnet-file>",
		Short: "upload a snapshot to preview the rendered file",
		Args:  cli.ArgsAny(),
	}
	opts := grizzlyOptsFromCmd(cmd)
	expires := cmd.Flags().IntP("expires", "e", 0, "when the preview should expire. Default 0 (never)")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) > 0 {
			opts.ResourceFile = &args[0]
		}
		resources, err := grizzly.Parse(registry, opts)
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
		Use:   "export <jsonnet-file> <dashboard-dir>",
		Short: "render Jsonnet and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	opts := grizzlyOptsFromCmd(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) > 0 {
			opts.ResourceFile = &args[0]
		}
		dashboardDir := args[1]
		resources, err := grizzly.Parse(registry, opts)
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

func grizzlyOptsFromCmd(cmd *cli.Command) grizzly.GrizzlyOpts {
	return grizzly.GrizzlyOpts{
		Directory:    cmd.Flags().StringP("directory", "d", "", "directory containing resource files"),
		Targets:      cmd.Flags().StringSliceP("target", "t", nil, "resources to target"),
		JsonnetPaths: cmd.Flags().StringSliceP("jpath", "J", getDefaultJsonnetFolders(), "Specify an additional library search dir (right-most wins)"),
	}
}

func getDefaultJsonnetFolders() []string {
	return []string{"vendor", "lib", "."}
}
