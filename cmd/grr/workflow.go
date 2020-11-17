package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grizzly"
)

func getCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "get <resource-type>.<resource-uid>",
		Short: "retrieve resource",
		Args:  cli.ArgsExact(1),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		return grizzly.Get(config, uid)
	}
	return cmd
}

func listCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "list <jsonnet-file>",
		Short: "list resource keys from file",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		resources, err := grizzly.Parse(config, jsonnetFile, *targets)
		if err != nil {
			return err
		}

		return grizzly.List(config, resources)
	}
	return cmd
}

func showCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "show <jsonnet-file>",
		Short: "render Jsonnet as json",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		resources, err := grizzly.Parse(config, jsonnetFile, *targets)
		if err != nil {
			return err
		}
		return grizzly.Show(config, resources)
	}
	return cmd
}

func diffCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <jsonnet-file>",
		Short: "compare Jsonnet resources with endpoint(s)",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		resources, err := grizzly.Parse(config, jsonnetFile, *targets)
		if err != nil {
			return err
		}
		return grizzly.Diff(config, resources)
	}
	return cmd
}

func applyCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <jsonnet-file>",
		Short: "render Jsonnet and push dashboard(s) to Grafana",
		Args:  cli.ArgsAny(),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		for _, jsonnetFile := range args {
			resources, err := grizzly.Parse(config, jsonnetFile, *targets)
			if err != nil {
				return err
			}
			grizzly.Apply(config, resources)
		}
		return nil
	}
	return cmd
}

type jsonnetWatchParser struct {
	jsonnetFile string
	targets     []string
}

func (p *jsonnetWatchParser) Name() string {
	return p.jsonnetFile
}

func (p *jsonnetWatchParser) Parse(config grizzly.Config) (grizzly.Resources, error) {
	return grizzly.Parse(config, p.jsonnetFile, p.targets)

}
func watchCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <jsonnet-file>",
		Short: "watch for file changes and apply",
		Args:  cli.ArgsExact(2),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		parser := &jsonnetWatchParser{
			jsonnetFile: args[1],
			targets:     *targets,
		}
		watchDir := args[0]

		return grizzly.Watch(config, watchDir, parser)

	}
	return cmd
}

func listenCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "listen <uid-to-watch> <output-file>",
		Short: "listen for file changes on remote and save locally",
		Args:  cli.ArgsExact(2),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		filename := args[1]
		return grizzly.Listen(config, uid, filename)
	}
	return cmd
}

func previewCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "preview <jsonnet-file>",
		Short: "upload a snapshot to preview the rendered file",
		Args:  cli.ArgsAny(),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Flags().IntP("expires", "e", 0, "when the preview should expire. Default 0 (never)")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		resources, err := grizzly.Parse(config, jsonnetFile, *targets)
		if err != nil {
			return err
		}
		e, err := cmd.Flags().GetInt("expires")
		if err != nil {
			return err
		}
		opts := &grizzly.PreviewOpts{
			ExpiresSeconds: e,
		}

		return grizzly.Preview(config, resources, opts)
	}
	return cmd
}

func exportCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "export <jsonnet-file> <dashboard-dir>",
		Short: "render Jsonnet and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "resources to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		dashboardDir := args[1]
		resources, err := grizzly.Parse(config, jsonnetFile, *targets)
		if err != nil {
			return err
		}
		return grizzly.Export(config, dashboardDir, resources)
	}
	return cmd
}

func providersCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "providers",
		Short: "Lists all providers registered with Grizzly",
		Args:  cli.ArgsExact(0),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		f := "%s\t%s\t%s\n"
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(w, f, "PROVIDER", "HANDLER", "JSON PATH")
		for _, provider := range config.Registry.Providers {
			for _, handler := range provider.GetHandlers() {
				for _, path := range handler.GetJSONPaths() {
					fmt.Fprintf(w, f, provider.GetName(), handler.GetName(), "/"+path)
				}
			}
		}
		return w.Flush()
	}
	return cmd
}
