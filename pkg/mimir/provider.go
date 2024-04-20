package mimir

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	config *config.MimirConfig
}

// NewProvider instantiates a new Provider.
func NewProvider(config *config.MimirConfig) (*Provider, error) {
	if _, err := exec.LookPath("cortextool"); err != nil {
		return nil, err
	}
	if config.Address == "" {
		return nil, fmt.Errorf("mimir address is not set")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("mimir api key is not set")
	}

	return &Provider{
		config: config,
	}, nil
}

func (p *Provider) Name() string {
	return "Mimir"
}

// Group returns the group name of the Grafana provider
func (p *Provider) Group() string {
	return "grizzly.grafana.com"
}

// Version returns the version of this provider
func (p *Provider) Version() string {
	return "v1alpha1"
}

// APIVersion returns the group and version of this provider
func (p *Provider) APIVersion() string {
	return filepath.Join(p.Group(), p.Version())
}

// GetHandlers identifies the handlers for the Grafana provider
func (p *Provider) GetHandlers() []grizzly.Handler {
	return []grizzly.Handler{
		NewRuleHandler(p),
	}
}
