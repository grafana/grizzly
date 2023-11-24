package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v3"
)

func Exists() (bool, error) {
	configPath := configdir.LocalConfig("grizzly")
	err := configdir.MakePath(configPath)
	if err != nil {
		return false, err
	}
	configFile := filepath.Join(configPath, "settings.yaml")
	_, err = os.Stat(configFile)
	return os.IsNotExist(err), nil
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
