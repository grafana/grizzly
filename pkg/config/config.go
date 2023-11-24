package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly/notifier"
	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v3"
)

func Init() error {
	exists, err := Exists()
	if err != nil {
		return err
	}
	if exists {
		notifier.Warn(nil, "Configuration already exists")
	}
	conf := Config{
		ApiVersion: "v1alpha1",
		Contexts: []Context{
			{
				Name: "default",
			},
		},
		CurrentContext: "default",
	}
	return Save(&conf)
}

func Exists() (bool, error) {
	configFile, err := configPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(configFile)
	return !os.IsNotExist(err), nil
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

func Load() (*Config, error) {
	configFile, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		config := Config{}
		Save(&config)
	}

	fh, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	var config Config
	decoder := yaml.NewDecoder(fh)
	decoder.Decode(&config)
	return &config, nil
}

func Save(config *Config) error {
	configFile, err := configPath()
	if err != nil {
		return err
	}

	fh, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	encoder := yaml.NewEncoder(fh)
	return encoder.Encode(config)
}

func GetContexts() error {
	conf, err := Load()
	if err != nil {
		return err
	}
	for _, context := range conf.Contexts {
		fmt.Println(context.Name)
	}
	return nil
}

func UseContext(context string) error {
	conf, err := Load()
	if err != nil {
		return err
	}
	conf.CurrentContext = context
	return Save(conf)
}

func CurrentContext() error {
	conf, err := Load()
	if err != nil {
		return err
	}
	fmt.Println(conf.CurrentContext)
	return nil
}

func Set(path string, value string) error {
	parts := strings.SplitN(path, ".", 2)
	y := fmt.Sprintf("%s:\n  %s: '%s'", parts[0], parts[1], value)

	conf, err := Load()
	if err != nil {
		return err
	}
	for i, context := range conf.Contexts {
		if context.Name == conf.CurrentContext {
			before, _ := yaml.Marshal(context)
			yaml.Unmarshal([]byte(y), &context)
			after, _ := yaml.Marshal(context)
			log.Printf("BEFORE: %s\nAFTER: %s\n", string(before), string(after))
			if string(before) == string(after) {
				return fmt.Errorf("Setting %s not recognised", path)
			}
			conf.Contexts[i] = context
			notifier.Info(nil, fmt.Sprintf("Setting %s set to %s", path, value))
			err = Save(conf)
			return err
		}
	}
	return fmt.Errorf("Current context %s not found", conf.CurrentContext)
}

func CreateContext(name string) error {
	conf, err := Load()
	if err != nil {
		return err
	}
	context := conf.GetContext(name)
	if context != nil {
		return fmt.Errorf("Context exists")
	}
	context = &Context{
		Name: name,
	}
	conf.Contexts = append(conf.Contexts, *context)
	conf.CurrentContext = name

	err = Save(conf)
	return err
}

func FromEnvironment() (*Config, error) {
	grafanaURL, exists := os.LookupEnv("GRAFANA_URL")
	if !exists {
		return nil, fmt.Errorf("Please configure Grizzly using grr config")
	}
	token, _ := os.LookupEnv("GRAFANA_TOKEN")
	user, _ := os.LookupEnv("GRAFANA_USER")

	grafanaConfig := GrafanaConfig{
		URL:   grafanaURL,
		Token: token,
		User:  user,
	}

	apiToken, _ := os.LookupEnv("GRAFANA_SM_TOKEN")
	stackID, _ := strconv.Atoi(os.Getenv("GRAFANA_SM_STACK_ID"))
	metricsInstanceID, _ := strconv.Atoi(os.Getenv("GRAFANA_SM_METRICS_ID"))
	logsInstanceID, _ := strconv.Atoi(os.Getenv("GRAFANA_SM_LOGS_ID"))
	var syntheticMonitoringConfig SyntheticMonitoringConfig
	if apiToken != "" {
		syntheticMonitoringConfig = SyntheticMonitoringConfig{
			Token:     apiToken,
			StackID:   int64(stackID),
			MetricsID: int64(metricsInstanceID),
			LogsID:    int64(logsInstanceID),
		}
	}
	conf := Config{
		Contexts: []Context{
			{
				Name:                "default",
				Grafana:             grafanaConfig,
				SyntheticMonitoring: syntheticMonitoringConfig,
			},
		},
		CurrentContext: "default",
	}
	return &conf, nil
}
