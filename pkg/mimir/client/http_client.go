package client

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/grafana/grizzly/pkg/mimir"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/grafana/grizzly/pkg/config"
)

var loadRulesEndpoint = "%s/prometheus/config/v1/rules/%s"
var listRulesEndpoint = "%s/prometheus/api/v1/alerts"

type Client struct {
	client *http.Client
	config *config.MimirConfig
}

func NewHttpClient(config *config.MimirConfig) (*Client, error) {
	client, err := createHttpClient()
	if err != nil {
		return nil, err
	}

	return &Client{config: config, client: client}, nil
}

func (c *Client) ListRules() (map[string][]mimir.PrometheusRuleGroup, error) {
	url := fmt.Sprintf(listRulesEndpoint, c.config.Address)
	res, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var group map[string][]mimir.PrometheusRuleGroup
	if err := yaml.Unmarshal(res, &group); err != nil {
		return nil, err
	}

	return group, nil
}

func (c *Client) LoadRules(resource mimir.PrometheusRuleGrouping) (string, error) {
	url := fmt.Sprintf(loadRulesEndpoint, c.config.Address, resource.Namespace)
	out, err := yaml.Marshal(resource.Groups)
	if err != nil {
		return "", fmt.Errorf("cannot marshall groups: %s", err)
	}

	res, err := c.doRequest(http.MethodPost, url, out)
	return string(res), err
}

func (c *Client) doRequest(method string, url string, body []byte) ([]byte, error) {
	if c.config.ApiKey == "" {
		return nil, errors.New("missing Mimir's APIKey")
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/yaml")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.ApiKey))
	if c.config.TenantID != 0 {
		req.Header.Set("X-Scope-OrgID", fmt.Sprintf("%d", c.config.TenantID))
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to load rules failed: %s", err)
	}

	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("error loading rules: %d", res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %s", err)
	}

	return b, nil
}

func createHttpClient() (*http.Client, error) {
	timeout := 10 * time.Second
	if timeoutStr := os.Getenv("GRIZZLY_HTTP_TIMEOUT"); timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, err
		}
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	return &http.Client{Timeout: timeout}, nil
}
