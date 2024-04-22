package client

import (
	"bytes"
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

func NewHttpClient(config *config.MimirConfig) Mimir {
	return &Client{config: config}
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
	if c.config.ApiKey != "" {
		req.SetBasicAuth(c.config.TenantID, c.config.ApiKey)
	} else {
		req.Header.Set("X-Scope-OrgID", fmt.Sprintf("%s", c.config.TenantID))
	}

	client, err := createHttpClient()
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
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
	// TODO: Move this configuration to the global configuration
	if timeoutStr := os.Getenv("GRIZZLY_HTTP_TIMEOUT"); timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, err
		}
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	return &http.Client{Timeout: timeout}, nil
}
