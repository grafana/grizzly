package grafana

import (
	"encoding/base64"
	"fmt"
	"net/http/httputil"
	"net/url"
	"path/filepath"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	config *config.GrafanaConfig
	client *gclient.GrafanaHTTPAPI
}

type ClientProvider interface {
	Client() (*gclient.GrafanaHTTPAPI, error)
	Config() *config.GrafanaConfig
}

// NewProvider instantiates a new Provider.
func NewProvider(config *config.GrafanaConfig) *Provider {
	return &Provider{
		config: config,
	}
}

func (p *Provider) Name() string {
	return "Grafana"
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
	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p.client != nil {
		return p.client, nil
	}

	parsedUrl, err := url.Parse(p.config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid Grafana URL")
	}

	transportConfig := gclient.DefaultTransportConfig().
		WithHost(parsedUrl.Host).
		WithSchemes([]string{parsedUrl.Scheme}).
		WithBasePath(filepath.Join(parsedUrl.Path, "api"))

	if p.config.Token != "" {
		if p.config.User != "" {
			transportConfig.BasicAuth = url.UserPassword(p.config.User, p.config.Token)
		} else {
			transportConfig.APIKey = p.config.Token
		}
	}
	grafanaClient := gclient.NewHTTPClientWithConfig(nil, transportConfig)
	p.client = grafanaClient
	return grafanaClient, nil
}

func (p *Provider) Config() *config.GrafanaConfig {
	return p.config
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
		NewAlertRuleGroupHandler(p),
		NewAlertNotificationPolicyHandler(p),
		NewAlertContactPointHandler(p),
	}
}

func (p *Provider) SetupProxy() (*httputil.ReverseProxy, error) {
	u, err := url.Parse(p.config.URL)
	if err != nil {
		return nil, err
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)

			if p.config.User != "" {
				header := fmt.Sprintf("%s:%s", p.config.User, p.config.Token)
				encoded := base64.StdEncoding.EncodeToString([]byte(header))
				r.Out.Header.Set("Authorization", "Bearer "+encoded)
			} else {
				r.Out.Header.Set("Authorization", "Bearer "+p.config.Token)
			}

			r.Out.Header.Del("Origin")
			r.Out.Header.Set("User-Agent", "Grizzly Proxy Server")
		},
	}
	return proxy, nil
}

func (p *Provider) Validate() error {
	if p.config.URL == "" {
		return fmt.Errorf("grafana URL is not set")
	}

	return nil
}
