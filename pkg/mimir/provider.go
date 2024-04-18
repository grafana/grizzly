package mimir

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir/client"
)

const (
	Http string = "http"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	config     *config.MimirConfig
	clientTool client.Mimir
}

// NewProvider instantiates a new Provider.
func NewProvider(config *config.MimirConfig) (*Provider, error) {
	clientTool := client.NewHttpClient(config)
	if config.Address == "" {
		return nil, fmt.Errorf("mimir address is not set")
	}
	if config.TenantID == "" {
		return nil, fmt.Errorf("mimir tenant id is not set")
	}

	return &Provider{
		config:     config,
		clientTool: clientTool,
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
		NewRuleHandler(p, p.clientTool),
	}
}
