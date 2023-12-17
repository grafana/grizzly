package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"
)

const (
	API_VERSION     = "v1alpha1"
	CURRENT_CONTEXT = "currentContext"
)

func Initialise() {
	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(configdir.LocalConfig("grizzly"))
	viper.AddConfigPath(".")

	viper.BindEnv("overrides.grafana.url", "GRAFANA_URL")
	viper.BindEnv("overrides.grafana.user", "GRAFANA_USER")
	viper.BindEnv("overrides.grafana.token", "GRAFANA_TOKEN")

	viper.BindEnv("overrides.synthetic-monitoring.token", "GRAFANA_SM_TOKEN")
	viper.BindEnv("overrides.synthetic-monitoring.stack-id", "GRAFANA_SM_STACK_ID")
	viper.BindEnv("overrides.synthetic-monitoring.logs-id", "GRAFANA_SM_METRICS_ID")
	viper.BindEnv("overrides.synthetic-monitoring.metrics-id", "GRAFANA_SM_LOGS_ID")

	viper.BindEnv("overrides.mimir.address", "CORTEX_ADDRESS")
	viper.BindEnv("overrides.mimir.tenant-id", "CORTEX_TENANT_ID")
	viper.BindEnv("overrides.mimir.api-key", "CORTEX_API_KEY")
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
	contextPath := fmt.Sprintf("contexts.%s", name)
	ctx := viper.Sub(contextPath)
	overrides := viper.Sub("overrides")
	if overrides != nil {
		for key, value := range overrides.AllSettings() {
			ctx.Set(key, value)
		}
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
	viper.Set("currentContext", "default")
	viper.Set("contexts.default.name", "default")
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
	contextPath := fmt.Sprintf("contexts.%s", name)
	ctx := viper.Sub(contextPath)
	overrides := viper.Sub("overrides")
	if overrides != nil {
		for key, value := range overrides.AllSettings() {
			ctx.Set(key, value)
		}
	}
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
	configpath, err := configPath()
	if err != nil {
		return err
	}
	writeConfigs := viper.New()
	for key, value := range viper.AllSettings() {
		if !strings.HasPrefix(key, "overrides") {
			writeConfigs.Set(key, value)
		}
	}
	err = writeConfigs.WriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			NewConfig()
			return writeConfigs.WriteConfigAs(configpath)
		}
	}
	return err
}
