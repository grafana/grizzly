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
		Use:   "get <dashboard-uid>",
		Short: "retrieve dashboard json",
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
		Short: "list dashboard keys from file",
		Args:  cli.ArgsExact(1),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]

		return grizzly.List(config, jsonnetFile)
	}
	return cmd
}

func showCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "show <jsonnet-file>",
		Short: "render Jsonnet dashboard as json",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		return grizzly.Show(config, jsonnetFile, *targets)
	}
	return cmd
}

func diffCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <jsonnet-file>",
		Short: "compare Jsonnet with dashboard(s) in Grafana",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		return grizzly.Diff(config, jsonnetFile, *targets)
	}
	return cmd
}

func applyCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <jsonnet-file>",
		Short: "render Jsonnet and push dashboard(s) to Grafana",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		return grizzly.Apply(config, jsonnetFile, *targets)
	}
	return cmd
}

func watchCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <jsonnet-file>",
		Short: "watch for file changes and apply",
		Args:  cli.ArgsExact(2),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		watchDir := args[0]
		jsonnetFile := args[1]
		return grizzly.Watch(config, watchDir, jsonnetFile, *targets)

	}
	return cmd
}

func previewCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "preview <jsonnet-file>",
		Short: "upload a snapshot to preview the rendered file",
		Args:  cli.ArgsAny(),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Flags().IntP("expires", "e", 0, "when the preview should expire. Default 0 (never)")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		e, err := cmd.Flags().GetInt("expires")
		if err != nil {
			return err
		}
		opts := &grizzly.PreviewOpts{
			ExpiresSeconds: e,
		}

		return grizzly.Preview(config, jsonnetFile, *targets, opts)
	}
	return cmd
}

func exportCmd(config grizzly.Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "export <jsonnet-file> <dashboard-dir>",
		Short: "render Jsonnet and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		dashboardDir := args[1]
		return grizzly.Export(config, jsonnetFile, dashboardDir, *targets)
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
		f := "%s\t%s\n"
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(w, f, "PROVIDER", "JSON PATH")
		for _, provider := range config.Registry.GetProviders() {
			path := provider.GetJSONPath()
			fmt.Fprintf(w, f, provider.GetName(), "/"+path)
		}
		return w.Flush()
	}
	return cmd
}
