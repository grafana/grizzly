package grafana

import (
	"fmt"
	"net/url"
	"path/filepath"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	context *config.Context
	client  *gclient.GrafanaHTTPAPI
}

type ClientProvider interface {
	Client() (*gclient.GrafanaHTTPAPI, error)
}

// NewProvider instantiates a new Provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Group returns the group name of the Grafana provider
func (p *Provider) Group() string {
	return "grizzly.grafana.com"
}

// Version returns the version of this provider
func (p *Provider) Version() string {
	return "v1alpha1"
}

func (p *Provider) Client() (*gclient.GrafanaHTTPAPI, error) {
	if p.client != nil {
		return p.client, nil
	}

	ctx, err := config.CurrentContext()
	if err != nil {
		return nil, err
	}
	p.context = ctx
	parsedUrl, err := url.Parse(p.context.Grafana.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid Grafana URL")
	}

	transportConfig := gclient.DefaultTransportConfig().
		WithHost(parsedUrl.Host).
		WithSchemes([]string{parsedUrl.Scheme})

	if p.context.Grafana.Token != "" {
		if p.context.Grafana.User != "" {
			transportConfig.BasicAuth = url.UserPassword(p.context.Grafana.User, p.context.Grafana.Token)
		} else {
			transportConfig.APIKey = p.context.Grafana.Token
		}
	}
	grafanaClient := gclient.NewHTTPClientWithConfig(nil, transportConfig)
	p.client = grafanaClient
	return grafanaClient, nil
}

// APIVersion returns the group and version of this provider
func (p *Provider) APIVersion() string {
	return filepath.Join(p.Group(), p.Version())
}

// GetHandlers identifies the handlers for the Grafana provider
func (p *Provider) GetHandlers() []grizzly.Handler {
	return []grizzly.Handler{
		NewDatasourceHandler(p),
		NewFolderHandler(p),
		NewLibraryElementHandler(p),
		NewDashboardHandler(p),
		NewRuleHandler(p),
		NewSyntheticMonitoringHandler(p),
	}
}
