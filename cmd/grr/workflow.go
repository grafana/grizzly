package main

import (
	"github.com/go-clix/cli"
	"github.com/malcolmholmes/grizzly/pkg/grizzly"
)

func getCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get <dashboard-uid>",
		Short: "retrieve dashboard json",
		Args:  cli.ArgsExact(1),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		uid := args[0]
		config, err := grizzly.ParseEnvironment()
		if err != nil {
			return err
		}
		return grizzly.Get(*config, uid)
	}
	return cmd
}

func listCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "list <jsonnet-file>",
		Short: "list dashboard keys from file",
		Args:  cli.ArgsExact(1),
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]

		return grizzly.List(jsonnetFile)
	}
	return cmd
}

func showCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "show <jsonnet-file>",
		Short: "render Jsonnet dashboard as json",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		config, err := grizzly.ParseEnvironment()
		if err != nil {
			return err
		}
		return grizzly.Show(*config, jsonnetFile, targets)
	}
	return cmd
}

func diffCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <jsonnet-file>",
		Short: "compare Jsonnet with dashboard(s) in Grafana",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		config, err := grizzly.ParseEnvironment()
		if err != nil {
			return err
		}
		return grizzly.Diff(*config, jsonnetFile, targets)
	}
	return cmd
}

func applyCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <jsonnet-file>",
		Short: "render Jsonnet and push dashboard(s) to Grafana",
		Args:  cli.ArgsExact(1),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		config, err := grizzly.ParseEnvironment()
		if err != nil {
			return err
		}
		return grizzly.Apply(*config, jsonnetFile, targets)
	}
	return cmd
}

func watchCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <jsonnet-file>",
		Short: "watch for file changes and apply",
		Args:  cli.ArgsExact(2),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		watchDir := args[0]
		jsonnetFile := args[1]
		config, err := grizzly.ParseEnvironment()
		if err != nil {
			return err
		}

		return grizzly.Watch(*config, watchDir, jsonnetFile, targets)

	}
	return cmd
}

func exportCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "export <jsonnet-file> <dashboard-dir>",
		Short: "render Jsonnet and save to a directory",
		Args:  cli.ArgsExact(2),
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		dashboardDir := args[1]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Export(*config, jsonnetFile, dashboardDir, targets)
	}
	return cmd
}
