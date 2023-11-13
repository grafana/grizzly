package grafana

import (
	"path/filepath"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	client *gclient.GrafanaHTTPAPI
}

// NewProvider instantiates a new Provider.
func NewProvider(client *gclient.GrafanaHTTPAPI) *Provider {
	return &Provider{
		client: client,
	}
}

// Group returns the group name of the Grafana provider
func (p Provider) Group() string {
	return "grizzly.grafana.com"
}

// Version returns the version of this provider
func (p Provider) Version() string {
	return "v1alpha1"
}

// APIVersion returns the group and version of this provider
func (p Provider) APIVersion() string {
	return filepath.Join(p.Group(), p.Version())
}

// GetHandlers identifies the handlers for the Grafana provider
func (p Provider) GetHandlers() []grizzly.Handler {
	return []grizzly.Handler{
		NewDatasourceHandler(p),
		NewFolderHandler(p),
		NewDashboardHandler(p),
		NewRuleHandler(p),
		NewSyntheticMonitoringHandler(p),
	}
}
