package testutil

import (
	"github.com/grafana/grizzly/pkg/config"
)

func TestContext() *config.Context {
	ctx := config.Context{
		Grafana: config.GrafanaConfig{
			URL: "http://localhost:3001",
		},
		Mimir: config.MimirConfig{
			Address:  "http://localhost:9009",
			TenantID: "demo",
			APIKey:   "abc",
		},
	}
	return &ctx
}
