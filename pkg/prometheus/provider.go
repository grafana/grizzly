package prometheus

import "github.com/grafana/grizzly/pkg/grizzly"

// Provider defines a Cortex Provider
type Provider struct{}

// NewProvider returns a new Cortex Provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetName returns the name of the Cortex provider
func (p *Provider) GetName() string {
	return "prometheus"
}

// GetHandlers identifies the handlers for the Cortex provider
func (p *Provider) GetHandlers() []grizzly.Handler {
	return []grizzly.Handler{
		NewRuleHandler(*p),
	}
}
