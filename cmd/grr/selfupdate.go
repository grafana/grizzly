package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-clix/cli"
	"github.com/grafana/grizzly/internal/grizzly"
	"github.com/grafana/grizzly/pkg/config"
)

func selfUpdateCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "self-update",
		Short: "Self update Grizzly",
		Args:  cli.ArgsExact(0),
	}
	var opts LoggingOpts

	cmd.Run = func(cmd *cli.Command, args []string) error {
		updater := grizzly.NewSelfUpdater(http.DefaultClient)

		newVersion, err := updater.UpdateSelf(context.Background(), config.Version)
		if errors.Is(err, grizzly.ErrNextVersionIsMajorBump) {
			return fmt.Errorf("self-update aborted as the next version (%[1]s) is a major bump from the current one (%[2]s). Please update manually: https://github.com/grafana/grizzly/releases/tag/%[1]s", newVersion, config.Version)
		}
		if errors.Is(err, grizzly.ErrCurrentVersionIsLatest) {
			fmt.Printf("Current version is the latest: %s\n", config.Version)
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Printf("Successfully updated to version %s\n", newVersion)

		return nil
	}

	return initialiseLogging(cmd, &opts)
}
