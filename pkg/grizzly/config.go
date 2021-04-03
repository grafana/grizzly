package grizzly

import (
	"fmt"

	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/mitchellh/mapstructure"
)

// GrizzlyOpts contains options for all Grizzly commands
type GrizzlyOpts struct {
	ConfigFile   *string
	ResourceFile *string
	JsonnetPaths *[]string
	Targets      *[]string
}

func (o *GrizzlyOpts) ConsumeArguments(args []string) {
	if len(args) > 0 {
		o.ResourceFile = &args[0]
	}
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}

type OutboundSource struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

const ConfigKind = "GrizzlyConfig"

type Config struct {
	Outbound []OutboundSource `yaml:"outbound"`
}

func NewConfig(configResources Resources) (*Config, error) {
	if len(configResources) == 0 {
		return nil, fmt.Errorf("No config resources found")
	}
	if len(configResources) > 1 {
		return nil, fmt.Errorf("Only one config resource allowed")
	}
	configResource := configResources[0]
	if configResource.Kind() != ConfigKind {
		return nil, fmt.Errorf("Expected resource of type %s, got %s", ConfigKind, configResource.Kind())
	}
	var config Config

	err := mapstructure.Decode(configResource["spec"], &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type ConfigParser struct{}

// Parse parses a manifest object into a struct for this resource type
func (p *ConfigParser) Parse(m manifest.Manifest) (Resources, error) {
	resource := Resource(m)
	return Resources{resource}, nil
}
