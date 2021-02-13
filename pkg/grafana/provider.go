package grafana

import "github.com/grafana/grizzly/pkg/grizzly"

// Provider defines a Grafana Provider
type Provider struct{}

// GetName returns the name of the Grafana provider
func (p *Provider) GetName() string {
	return "grafana"
}

// GetHandlers identifies the handlers for the Grafana provider
func (p *Provider) GetHandlers() []grizzly.Handler {
	return []grizzly.Handler{
		NewDashboardHandler(*p),
		NewDatasourceHandler(*p),
		NewSyntheticMonitoringHandler(*p),
	}
}
