package syntheticmonitoring

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/grizzly"
	smapi "github.com/grafana/synthetic-monitoring-api-go-client"
)

var smAPIURLsExceptions = map[string]string{
	"au":              "https://synthetic-monitoring-api-au-southeast.grafana.net",
	"eu":              "https://synthetic-monitoring-api-eu-west.grafana.net",
	"prod-gb-south-0": "https://synthetic-monitoring-api-gb-south.grafana.net",
	"us":              "https://synthetic-monitoring-api.grafana.net",
	"us-azure":        "https://synthetic-monitoring-api-us-central2.grafana.net",
}

// Provider is a grizzly.Provider implementation for Grafana.
type Provider struct {
	config *config.SyntheticMonitoringConfig
}

type ClientProvider interface {
	Client() (*smapi.Client, error)
}

// NewProvider instantiates a new Provider.
func NewProvider(config *config.SyntheticMonitoringConfig) (*Provider, error) {
	if config.StackID == 0 {
		return nil, fmt.Errorf("stack id is not set")
	}
	if config.MetricsID == 0 {
		return nil, fmt.Errorf("metrics id is not set")
	}
	if config.LogsID == 0 {
		return nil, fmt.Errorf("logs id is not set")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("token is not set")
	}
	if config.Region == "" {
		return nil, fmt.Errorf("region is not set")
	}

	return &Provider{
		config: config,
	}, nil
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
	client, err := NewHttpClient()
	if err != nil {
		return nil, err
	}

	url := smAPIURLsExceptions[p.config.Region]
	if url == "" {
		url = fmt.Sprintf("https://synthetic-monitoring-api-%s.grafana.net", strings.TrimPrefix(p.config.Region, "prod-"))
	}

	smClient := smapi.NewClient(url, "", client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = smClient.Install(ctx, p.config.StackID, p.config.MetricsID, p.config.LogsID, p.config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to install synthetic monitoring client : %v", err)
	}

	return smClient, nil
}
