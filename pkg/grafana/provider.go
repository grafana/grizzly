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

var foo []grizzly.Handler

// GetHandlers identifies the handlers for the Grafana provider
func (p *Provider) GetHandlers() []grizzly.Handler {
	handlers := []grizzly.Handler{
		NewDatasourceHandler(*p),
		NewFolderHandler(*p),
		NewDashboardHandler(*p),
		NewRuleHandler(*p),
		NewSyntheticMonitoringHandler(*p),
	}

	for _, resourceKind := range p.ResourceKinds() {
		handlers = append(handlers, NewComposableHandler(*p, resourceKind))
	}
	return handlers
}

func (p *Provider) ResourceKinds() []grizzly.ResourceKind {
	return []grizzly.ResourceKind{
		{
			Kind:         "DashboardComposition",
			ResolvedKind: "Dashboard",
			References: []grizzly.Reference{
				{
					Kind: "Template",
					Path: "spec.templates",
					Name: "reference",
					Type: "array",
				},
				{
					Kind: "Panel",
					Path: "spec.panels",
					Name: "reference",
					Type: "array",
					Fields: []string{
						"gridPos",
					},
				},
			},
		}, {
			Kind: "Panel",
			References: []grizzly.Reference{
				{
					Kind:   "Query",
					Path:   "spec.targets",
					Name:   "reference",
					Type:   "array",
					Fields: []string{"format", "instant", "range"},
				},
				{
					Kind: "PanelDefaults",
					Path: "spec.defaults",
					Name: "reference",
					Type: "array",
				},
			},
		}, {
			Kind:   "PanelDefaults",
			AtRoot: true,
		}, {
			Kind:                "Query",
			InterpolationTarget: "query",
			Interpolations: []grizzly.Reference{
				{
					Kind: "Query",
					Path: "spec.variables",
					Name: "reference",
					Type: "array",
				},
			},
			Inputs: []grizzly.Reference{
				{
					Kind: "Query",
					Path: "spec.inputs",
					Name: "reference",
					Type: "array",
				},
			},
		}, {
			Kind: "Template",
		},
	}
}
