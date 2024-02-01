package testutil

import (
	"github.com/grafana/grizzly/pkg/config"
	"github.com/spf13/viper"
)

func GetUrl() string {
	return "http://localhost:3001/"
}

func InitialiseTestConfig() {
	viper.Set(config.CURRENT_CONTEXT, "test")
	viper.Set("contexts.test.grafana.name", "test")
	viper.Set("contexts.test.grafana.url", GetUrl())
}
