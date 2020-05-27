package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/go-clix/cli"
	"github.com/malcolmholmes/grafana-dash/pkg/dash"
)

func getCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "get <dashboard-uid>",
		Short: "retrieve dashboard json",
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

func listCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "list <jsonnet-file>",
		Short: "list dashboard keys from file",
	}
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]

		return dash.List(jsonnetFile)
	}
	return cmd
}

func showCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "show <jsonnet-file>",
		Short: "render Jsonnet dashboard as json",
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Show(*config, jsonnetFile, targets)
	}
	return cmd
}

func diffCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <jsonnet-file>",
		Short: "compare Jsonnet with dashboard(s) in Grafana",
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Diff(*config, jsonnetFile, targets)
	}
	return cmd
}

func applyCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <jsonnet-file>",
		Short: "render Jsonnet and push dashboard(s) to Grafana",
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		jsonnetFile := args[0]
		config, err := dash.ParseEnvironment()
		if err != nil {
			return err
		}
		return dash.Apply(*config, jsonnetFile, targets)
	}
	return cmd
}

func watchCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "watch <dir-to-watch> <jsonnet-file>",
		Short: "watch for file changes and apply",
	}
	targets := cmd.Flags().StringSliceP("target", "t", nil, "dashboards to target")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		watchFile := args[0]
		jsonnetFile := args[1]

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						config, err := dash.ParseEnvironment()
						if err != nil {
							log.Println("error:", err)
						}
						if err := dash.Apply(*config, jsonnetFile, targets); err != nil {
							log.Println("error:", err)
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()

		err = watcher.Add(watchFile)
		if err != nil {
			return err
		}
		<-done
		return nil
	}
	return cmd
}
