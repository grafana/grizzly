package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	CurrentContextSetting = "current-context"
)

// Version is the current version of the grr command.
// To be overwritten at build time
var Version = "dev"

func Initialise() {
	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath(configdir.LocalConfig("grizzly"))
}

func override(v *viper.Viper) {
	bindings := map[string]string{
		"grafana.url":   "GRAFANA_URL",
		"grafana.user":  "GRAFANA_USER",
		"grafana.token": "GRAFANA_TOKEN",

		"synthetic-monitoring.access-token": "GRAFANA_SM_ACCESS_TOKEN",
		"synthetic-monitoring.token":        "GRAFANA_SM_TOKEN",
		"synthetic-monitoring.stack-id":     "GRAFANA_SM_STACK_ID",
		"synthetic-monitoring.logs-id":      "GRAFANA_SM_LOGS_ID",
		"synthetic-monitoring.metrics-id":   "GRAFANA_SM_METRICS_ID",
		"synthetic-monitoring.url":          "GRAFANA_SM_URL",

		"mimir.address":    "MIMIR_ADDRESS",
		"mimir.tenant-id":  "MIMIR_TENANT_ID",
		"mimir.api-key":    "MIMIR_API_KEY",
		"mimir.auth-token": "MIMIR_AUTH_TOKEN",
	}

	// To keep retro compatibility
	legacyBindings := map[string]string{
		"MIMIR_ADDRESS":   "CORTEX_ADDRESS",
		"MIMIR_TENANT_ID": "CORTEX_TENANT_ID",
		"MIMIR_API_KEY":   "CORTEX_API_KEY",
	}

	for key, env := range bindings {
		if val := getVal(env, legacyBindings); val != "" {
			v.Set(key, val)
		}
	}
}

func getVal(env string, alternativeMap map[string]string) string {
	if val := os.Getenv(env); val != "" {
		return val
	}

	if alternativeMap[env] != "" {
		return getVal(alternativeMap[env], nil)
	}

	return ""
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
	name := viper.GetString(CurrentContextSetting)
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

func NewConfig() {
	viper.Set("apiVersion", "v1alpha1")
	viper.Set(CurrentContextSetting, "default")
	viper.Set("contexts.default.name", "default")
}

func GetContexts() ([]string, error) {
	contexts := map[string]interface{}{}
	if err := viper.UnmarshalKey("contexts", &contexts); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(contexts))
	for k := range contexts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func UseContext(context string) error {
	contexts := map[string]interface{}{}
	if err := viper.UnmarshalKey("contexts", &contexts); err != nil {
		return err
	}
	for k := range contexts {
		if k == context {
			viper.Set(CurrentContextSetting, context)
			return Write()
		}
	}
	return fmt.Errorf("context %s not found", context)
}

func CurrentContext() (*Context, error) {
	name := viper.GetString(CurrentContextSetting)
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
	if err := ctx.Unmarshal(&context); err != nil {
		return nil, err
	}
	context.Name = name
	return &context, nil
}

var acceptableKeys = map[string]string{
	"grafana.url":                       "string",
	"grafana.token":                     "string",
	"grafana.user":                      "string",
	"grafana.insecure-skip-verify":      "bool",
	"grafana.tls-host":                  "string",
	"mimir.address":                     "string",
	"mimir.tenant-id":                   "string",
	"mimir.api-key":                     "string",
	"mimir.auth-token":                  "string",
	"synthetic-monitoring.access-token": "string",
	"synthetic-monitoring.token":        "string",
	"synthetic-monitoring.stack-id":     "int",
	"synthetic-monitoring.metrics-id":   "int",
	"synthetic-monitoring.logs-id":      "int",
	"synthetic-monitoring.url":          "string",
	"targets":                           "[]string",
	"output-format":                     "string",
	"only-spec":                         "bool",
}

func Hash() (string, error) {
	cfg := viper.AllSettings()
	out := sha256.New()
	err := json.NewEncoder(out).Encode(cfg)
	if err != nil {
		return "", fmt.Errorf("writing to hash")
	}
	return fmt.Sprintf("%X", out.Sum(nil)), nil
}

func Get(path, outputFormat string) (string, error) {
	ctx := viper.GetString(CurrentContextSetting)

	vCtx := viper.Sub(fmt.Sprintf("contexts.%s", ctx))
	if vCtx == nil {
		vCtx = viper.New()
	}
	override(vCtx)

	var val any
	val = vCtx.AllSettings()

	if path != "" {
		for _, part := range strings.Split(path, ".") {
			if val == nil {
				break
			}

			values, ok := val.(map[string]interface{})
			if !ok {
				val = nil
				break
			}

			val, ok = values[part]
			if !ok {
				val = nil
				break
			}
		}
	}

	if val == nil {
		return "", fmt.Errorf("key not found: %s", path)
	}
	switch outputFormat {
	case "yaml":
		res, err := yaml.Marshal(val)
		return string(res), err
	case "json":
		res, err := json.MarshalIndent(val, "", "  ")
		return string(res), err
	}
	return "", fmt.Errorf("unknown output format: %s", outputFormat)
}

func Set(path string, value string) error {
	for key, typ := range acceptableKeys {
		if path == key {
			ctx := viper.GetString(CurrentContextSetting)
			fullPath := fmt.Sprintf("contexts.%s.%s", ctx, path)
			var val any
			switch typ {
			case "string":
				val = value
			case "[]string":
				val = strings.Split(value, ",")
			case "bool":
				val = strings.ToLower(value) == "true"
			case "int":
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("key %s should be an integer: %s", key, err)
				}
				val = intValue
			default:
				return fmt.Errorf("unknown config key type %s for key %s", typ, key)
			}
			viper.Set(fullPath, val)
			return Write()
		}
	}
	return fmt.Errorf("key not recognised: %s", path)
}

func Unset(path string) error {
	exists := false
	for k := range acceptableKeys {
		if path == k {
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("%s is not a valid path", path)
	}

	ctx := viper.GetString(CurrentContextSetting)
	fullPath := fmt.Sprintf("contexts.%s.%s", ctx, path)

	if !viper.InConfig(fullPath) {
		return fmt.Errorf("key %s is already unset", path)
	}

	parts := strings.Split(path, ".")
	allConfig := viper.AllSettings()
	deleteValue(allConfig, parts[1], "contexts", ctx, parts[0])

	encodedConfig, err := json.MarshalIndent(allConfig, "", "	")
	if err != nil {
		return fmt.Errorf("error encoding new configuration: %s", err)
	}

	if err = viper.ReadConfig(bytes.NewReader(encodedConfig)); err != nil {
		return fmt.Errorf("error reading new configuration: %s", err)
	}

	return viper.WriteConfig()
}

func deleteValue(settings map[string]any, deleteKey string, iteratorKeys ...string) {
	for _, k := range iteratorKeys {
		settings = settings[k].(map[string]any)
	}
	delete(settings, deleteKey)
}

func CreateContext(name string) error {
	viper.Set(CurrentContextSetting, name)
	viper.Set(fmt.Sprintf("contexts.%s.name", name), name)
	return Write()
}

func Write() error {
	err := viper.WriteConfig()
	if err == nil {
		return nil
	}

	// We only know how to handle `viper.ConfigFileNotFoundError` errors.
	// Everything else bubbles up.
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		return err
	}

	// Ensure that our configuration directory exists: viper only takes care of
	// creating the file.
	configDir := configdir.LocalConfig("grizzly")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0700); err != nil {
			return err
		}
	}

	// Viper failed because no configuration file exists in the "config path".
	// We explicitly tell it where to write its config: at the most global location.
	globalConfigPath := filepath.Join(configDir, "settings.yaml")

	return viper.WriteConfigAs(globalConfigPath)
}

func (c *Context) GetTargets(overrides []string) []string {
	if len(overrides) > 0 {
		return overrides
	}
	return c.Targets
}
