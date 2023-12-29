package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"
)

const (
	API_VERSION     = "v1alpha1"
	CURRENT_CONTEXT = "current-context"
)

func Initialise() {
	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(configdir.LocalConfig("grizzly"))
	viper.AddConfigPath(".")
}

func override(v *viper.Viper) {
	bindings := map[string]string{
		"grafana.url":   "GRAFANA_URL",
		"grafana.user":  "GRAFANA_USER",
		"grafana.token": "GRAFANA_TOKEN",

		"synthetic-monitoring.token":      "GRAFANA_SM_TOKEN",
		"synthetic-monitoring.stack-id":   "GRAFANA_SM_STACK_ID",
		"synthetic-monitoring.logs-id":    "GRAFANA_SM_METRICS_ID",
		"synthetic-monitoring.metrics-id": "GRAFANA_SM_LOGS_ID",

		"mimir.address":   "CORTEX_ADDRESS",
		"mimir.tenant-id": "CORTEX_TENANT_ID",
		"mimir.api-key":   "CORTEX_API_KEY",
	}
	for key, env := range bindings {
		val := os.Getenv(env)
		if val != "" {
			v.Set(key, val)
		}
	}
}

func Read() error {
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			NewConfig()
		} else {
			return err
		}
	}
	return nil
}

func Mock(values map[string]interface{}) {
	for k, v := range values {
		viper.Set(k, v)
	}
}

func Import() error {
	name := viper.GetString(CURRENT_CONTEXT)
	if name == "" {
		NewConfig()
		return Import()
	}
	contextPath := fmt.Sprintf("contexts.%s", name)
	ctx := viper.Sub(contextPath)
	if ctx == nil {
		ctx = viper.New()
	}
	override(ctx)
	for k, v := range ctx.AllSettings() {
		viper.Set(contextPath+"."+k, v)
	}
	err := Write()
	return err
}

func configPath() (string, error) {
	configPath := configdir.LocalConfig("grizzly")
	err := configdir.MakePath(configPath)
	if err != nil {
		return "", err
	}

	configFile := filepath.Join(configPath, "settings.yaml")
	return configFile, nil
}

func NewConfig() {
	viper.Set("apiVersion", "v1alpha1")
	viper.Set(CURRENT_CONTEXT, "default")
}

func GetContexts() error {
	contexts := map[string]interface{}{}
	viper.UnmarshalKey("contexts", &contexts)
	for k := range contexts {
		fmt.Printf("  %s\n", k)
	}
	return nil
}

func UseContext(context string) error {
	viper.Set(CURRENT_CONTEXT, context)
	return Write()
}

func CurrentContext() (*Context, error) {
	name := viper.GetString(CURRENT_CONTEXT)
	if name == "" {
		NewConfig()
		return CurrentContext()
	}
	contextPath := fmt.Sprintf("contexts.%s", name)
	ctx := viper.Sub(contextPath)
	if ctx == nil {
		ctx = viper.New()
	}
	override(ctx)
	var context Context
	ctx.Unmarshal(&context)
	context.Name = name
	return &context, nil
}

func Set(path string, value string) error {
	ctx := viper.GetString(CURRENT_CONTEXT)
	fullPath := fmt.Sprintf("contexts.%s.%s", ctx, path)
	viper.Set(fullPath, value)
	return Write()
}

func CreateContext(name string) error {
	viper.Set(CURRENT_CONTEXT, name)
	viper.Set("contexts.default.name", name)
	return Write()
}

func Write() error {
	err := viper.WriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configpath, err := configPath()
			if err != nil {
				return err
			}
			return viper.WriteConfigAs(configpath)
		}
	}
	return err
}
