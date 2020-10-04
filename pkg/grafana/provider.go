package grafana

import "github.com/grafana/grizzly/pkg/grizzly"

// Provider defines a Grafana Provider
type Provider struct{}

// NewProvider returns a new Grafana Provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetName returns the name of the Grafana provider
func (p *Provider) GetName() string {
	return "grafana"
}

// GetHandlers identifies the handlers for the Grafana provider
func (p *Provider) GetHandlers() []grizzly.Handler {
	return []grizzly.Handler{
		&DashboardHandler{},
		&DatasourceHandler{},
		//PluginHandler{},
		&SyntheticMonitoringHandler{},
		//MixinHandler{},
	}
}
