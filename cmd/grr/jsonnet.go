package main

import (
	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/pkg/grizzly"
)

func jsonnetCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "jsonnet <sub-command>",
		Short: "Jsonnet related commands",
		Args:  cli.ArgsExact(0),
	}
	cmd.AddCommand(jsonnetImportsCmd())
	return cmd
}

func jsonnetImportsCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "imports <dir> <imports-file>",
		Short: "Generate a Jsonnet file that imports all files in a directory",
		Args:  cli.ArgsExact(2),
	}
	var opts grizzly.Opts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		dir := args[0]
		out := args[1]
		return grizzly.GenerateJsonnetImports(dir, out)
	}

	cmd.Flags().StringVarP(&opts.FolderUID, "folder", "f", generalFolderUID, "folder to push dashboards to")
	cmd = initialiseOnlySpec(cmd, &opts)
	return initialiseCmd(cmd, &opts)
}
