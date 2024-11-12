package syntheticmonitoring

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
	smapi "github.com/grafana/synthetic-monitoring-api-go-client"
)

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	config *config.SyntheticMonitoringConfig
}

type ClientProvider interface {
	Client() (*smapi.Client, error)
}

// NewProvider instantiates a new Provider.
func NewProvider(config *config.SyntheticMonitoringConfig) *Provider {
	return &Provider{
		config: config,
	}
}

func (p *Provider) Validate() error {
	if p.config.URL == "" {
		p.config.URL = "https://synthetic-monitoring-api.grafana.net"
	}

	smInstallationConfigured := p.config.StackID != 0 && p.config.MetricsID != 0 && p.config.LogsID != 0 && p.config.Token != ""

	if p.config.AccessToken != "" && smInstallationConfigured {
		return fmt.Errorf("both access token and stack configuration (stack id, metrics id, logs id, token) are set. Only one can be used")
	}

	if p.config.AccessToken == "" && !smInstallationConfigured {
		return fmt.Errorf("neither access token nor stack configuration (stack id, metrics id, logs id, token) are set. One must be set")
	}

	return nil
}

func (p *Provider) Status() grizzly.ProviderStatus {
	status := grizzly.ProviderStatus{}

	if err := p.Validate(); err != nil {
		status.ActiveReason = err.Error()
		return status
	}

	status.Active = true

	client, err := p.Client()
	if err != nil {
		status.OnlineReason = err.Error()
		return status
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err = client.ListChecks(ctx); err != nil {
		status.OnlineReason = err.Error()
		return status
	}

	status.Online = true

	return status
}

func (p *Provider) Name() string {
	return "Synthetic Monitoring"
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
		NewSyntheticMonitoringHandler(p),
	}
}

// NewClient creates a new client for synthetic monitoring go client
func (p *Provider) Client() (*smapi.Client, error) {
	client, err := httputils.NewHTTPClient()
	if err != nil {
		return nil, err
	}

	if p.config.AccessToken != "" {
		smClient := smapi.NewClient(p.config.URL, p.config.AccessToken, client)
		return smClient, nil
	}

	smClient := smapi.NewClient(p.config.URL, "", client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = smClient.Install(ctx, p.config.StackID, p.config.MetricsID, p.config.LogsID, p.config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to install synthetic monitoring client : %v", err)
	}

	return smClient, nil
}
