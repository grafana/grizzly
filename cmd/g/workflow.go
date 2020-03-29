package main

import (
	"github.com/go-clix/cli"
	"github.com/malcolmholmes/grafana-dash/pkg/dash"
)
func getCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get <dashboard-uid>",
		Short: "retrieve dashboard json",
		Args:  cli.Args{},
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {

		uid := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Get(*config, uid)
	}
	return cmd
}

func showCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "show <jsonnet-file>",
		Short: "render Jsonnet dashboard as json",
		Args:  cli.Args{},
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {

		jsonnetFile := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Show(*config, jsonnetFile)
	}
	return cmd
}

func diffCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <jsonnet-file>",
		Short: "compare Jsonnet with dashboard(s) in Grafana",
		Args:  cli.Args{},
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {

		jsonnetFile := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Diff(*config, jsonnetFile)
	}
	return cmd
}

func applyCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <jsonnet-file>",
		Short: "render Jsonnet and push dashboard(s) to Grafana",
		Args:  cli.Args{},
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {

		jsonnetFile := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Apply(*config, jsonnetFile)
	}
	return cmd
}