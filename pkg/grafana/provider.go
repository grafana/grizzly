package grafana

import (
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
)

// Provider defines a Grafana Provider
type Provider struct{}

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
		NewFolderHandler(*p),
		NewDashboardHandler(*p),
		NewDatasourceHandler(*p),
		NewSyntheticMonitoringHandler(*p),
		NewRuleHandler(*p),
	}
}
