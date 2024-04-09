package client

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/mimir/models"
	"gopkg.in/yaml.v3"
)

type CortexTool struct {
	config *config.MimirConfig
}

func NewCortexTool(config *config.MimirConfig) Mimir {
	return &CortexTool{config: config}
}

func (c *CortexTool) ListRules() (map[string][]models.PrometheusRuleGroup, error) {
	cmd := exec.Command("cortextool", "rules", "print", "--disable-color")
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_TENANT_ID=%s", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_API_KEY=%s", c.config.ApiKey))
	res, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var group map[string][]models.PrometheusRuleGroup
	if err := yaml.Unmarshal(res, &group); err != nil {
		return nil, err
	}

	return group, nil
}

func (c *CortexTool) CreateRules(resource models.PrometheusRuleGrouping) (string, error) {
	tmpFile, err := createTmpFile(resource)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cmd := exec.Command("cortextool", "rules", "load", tmpFile)
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_TENANT_ID=%s", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_API_KEY=%s", c.config.ApiKey))
	res, err := cmd.Output()
	return string(res), err
}

func createTmpFile(resource models.PrometheusRuleGrouping) (string, error) {
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
