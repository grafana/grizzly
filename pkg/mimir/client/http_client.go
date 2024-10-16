package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/mimir/models"
	"gopkg.in/yaml.v3"
)

var loadRulesEndpoint = "%s/prometheus/config/v1/rules/%s"
var listRulesEndpoint = "%s/prometheus/api/v1/rules"
var alertmanagerAPIPath = "%s/api/v1/alerts"

type ListGroupResponse struct {
	Status string `yaml:"status"`
	Data   struct {
		DataGroups []DataGroups `yaml:"groups"`
	} `yaml:"data"`
}

type DataGroups struct {
	Name  string        `yaml:"name"`
	File  string        `yaml:"file"`
	Rules []interface{} `yaml:"rules"`
}

type Client struct {
	config *config.MimirConfig
}

func NewHTTPClient(config *config.MimirConfig) Mimir {
	return &Client{config: config}
}

func (c *Client) CreateAlertmangerConfig(resource models.PrometheusAlertmanagerConfig) error {
	url := fmt.Sprintf(alertmanagerAPIPath, c.config.Address)
	cfg, err := yaml.Marshal(&resource)
	if err != nil {
		return err
	}

	_, err = c.doRequest(http.MethodPost, url, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAlertmanagerConfig() (*models.PrometheusAlertmanagerConfig, error) {
	url := fmt.Sprintf(alertmanagerAPIPath, c.config.Address)
	res, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	config := models.PrometheusAlertmanagerConfig{}
	err = yaml.Unmarshal(res, &config)
	if err != nil {
		return nil, errors.New("unable to unmarshal response")
	}

	return &config, nil
}

func (c *Client) ListRules() (map[string][]models.PrometheusRuleGroup, error) {
	url := fmt.Sprintf(listRulesEndpoint, c.config.Address)
	res, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response ListGroupResponse
	if err := yaml.Unmarshal(res, &response); err != nil {
		return nil, err
	}

	groups := make(map[string][]models.PrometheusRuleGroup)
	for _, g := range response.Data.DataGroups {
		groups[g.File] = append(groups[g.File], models.PrometheusRuleGroup{
			Name:  g.Name,
			Rules: g.Rules,
		})
	}

	return groups, nil
}

func (c *Client) CreateRules(resource models.PrometheusRuleGrouping) error {
	url := fmt.Sprintf(loadRulesEndpoint, c.config.Address, resource.Namespace)
	for _, group := range resource.Groups {
		out, err := yaml.Marshal(group)
		if err != nil {
			return fmt.Errorf("cannot marshall groups: %s", err)
		}

		if _, err = c.doRequest(http.MethodPost, url, out); err != nil {
			return fmt.Errorf("error found creating rule group: %s", group.Name)
		}
	}

	return nil
}

func (c *Client) doRequest(method string, url string, body []byte) ([]byte, error) {
	if c.config.TenantID == "" {
		return nil, errors.New("missing tenant-id")
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/yaml")
	switch {
	case c.config.APIKey != "":
		req.SetBasicAuth(c.config.TenantID, c.config.APIKey)
	case c.config.AuthToken != "":
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	default:
		req.Header.Set("X-Scope-OrgID", c.config.TenantID)
	}

	client, err := c.createHTTPClient()
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %s", err)
	}

	if res.StatusCode >= 300 && res.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("error: %d", res.StatusCode)
	} else if res.StatusCode == http.StatusNotFound {
		return []byte(nil), nil
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %s", err)
	}

	return b, nil
}

func (c *Client) createHTTPClient() (*http.Client, error) {
	timeout := 10 * time.Second
	// TODO: Move this configuration to the global configuration
	if timeoutStr := os.Getenv("GRIZZLY_HTTP_TIMEOUT"); timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, err
		}
		timeout = time.Duration(timeoutSeconds) * time.Second
	}

	tlsConfig := &tls.Config{}
	httpClient := http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	if c.config.TLS.CAPath != "" {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		caCertPEM, err := os.ReadFile(c.config.TLS.CAPath)
		if err != nil {
			return nil, err
		}

		ok := certPool.AppendCertsFromPEM(caCertPEM)
		if !ok {
			return nil, fmt.Errorf("could not append ca-bundle at path %s to existing certificates", c.config.TLS.CAPath)
		}

		tlsConfig.RootCAs = certPool
	}

	if c.config.TLS.ClientCertPath != "" || c.config.TLS.ClientKeyPath != "" {
		clientTLSCert, err := tls.LoadX509KeyPair(c.config.TLS.ClientCertPath, c.config.TLS.ClientKeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientTLSCert}
	}

	httpClient.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return &httpClient, nil
}
