package mimir

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir/client"
)

const (
	MimirTool  string = "mimirtool"
	CortexTool string = "cortextool"
	Http       string = "http"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	config     *config.MimirConfig
	clientTool client.Mimir
}

// NewProvider instantiates a new Provider.
func NewProvider(config *config.MimirConfig) (*Provider, error) {
	var clientTool client.Mimir
	switch config.Client {
	case MimirTool:
		if !isBinarySet(config.MimirToolPath, MimirTool) {
			return nil, ErrNoBinarySet{name: MimirTool}
		}
		clientTool = client.NewMimirTool(config)
	case Http:
		clientTool = client.NewHttpClient(config)
	default:
		// Uses cortextool as we were using from the beginning.
		if !isBinarySet(config.CortexToolPath, CortexTool) {
			return nil, ErrNoBinarySet{name: CortexTool}
		}
		clientTool = client.NewCortexTool(config)
	}

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

func isBinarySet(path string, tool string) bool {
	if path != "" {
		return true
	}

	_, err := exec.LookPath(tool)
	if err != nil {
		return false
	}

	return true
}
