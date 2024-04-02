package syntheticmonitoring

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

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

// Client creates a new client for synthetic monitoring go client
func (p *Provider) Client() (*smapi.Client, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}

	client, err := NewHttpClient()
	if err != nil {
		return nil, err
	}

	smClient := smapi.NewClient(smBaseURL, "", client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = smClient.Install(ctx, p.config.StackID, p.config.MetricsID, p.config.LogsID, p.config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to install synthetic monitoring client : %v", err)
	}

	return smClient, nil
}

func (p *Provider) validate() error {
	if p.config.StackID == 0 {
		return fmt.Errorf("stack id is not set")
	}
	if p.config.MetricsID == 0 {
		return fmt.Errorf("metrics id is not set")
	}
	if p.config.LogsID == 0 {
		return fmt.Errorf("logs id is not set")
	}
	if p.config.Token == "" {
		return fmt.Errorf("token is not set")
	}

	return nil
}
