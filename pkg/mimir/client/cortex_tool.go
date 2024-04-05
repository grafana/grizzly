package client

import (
	"fmt"
	"github.com/grafana/grizzly/pkg/mimir"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"

	"github.com/grafana/grizzly/pkg/config"
)

type CortexTool struct {
	config *config.MimirConfig
}

func NewCortexTool(config *config.MimirConfig) *CortexTool {
	return &CortexTool{config: config}
}

func (c *CortexTool) ListRules() (map[string][]mimir.PrometheusRuleGroup, error) {
	cmd := exec.Command("rules", "print", "--disable-color")
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_TENANT_ID=%d", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_API_KEY=%s", c.config.ApiKey))
	res, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var group map[string][]mimir.PrometheusRuleGroup
	if err := yaml.Unmarshal(res, &group); err != nil {
		return nil, err
	}

	return group, nil
}

func (c *CortexTool) LoadRules(resource mimir.PrometheusRuleGrouping) (string, error) {
	tmpFile, err := createTmpFile(resource)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cmd := exec.Command("rules", "load", tmpFile)
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_TENANT_ID=%d", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_API_KEY=%s", c.config.ApiKey))
	res, err := cmd.Output()
	return string(res), err
}

func createTmpFile(resource mimir.PrometheusRuleGrouping) (string, error) {
	tmpfile, err := os.CreateTemp("", "cortextool-*")
	if err != nil {
		return "", err
	}

	out, err := yaml.Marshal(resource)
	if err != nil {
		return "", err
	}

	os.WriteFile(tmpfile.Name(), out, 0644)
	return tmpfile.Name(), nil
}
