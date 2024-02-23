package testutil

import (
	"github.com/grafana/grizzly/pkg/config"
)

func TestContext() *config.Context {
	ctx := config.Context{
		Grafana: config.GrafanaConfig{
			URL: "http://localhost:3001",
		},
	}
	return &ctx
}
