package client

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/mimir/models"
	"gopkg.in/yaml.v3"
)

type MimirTool struct {
	config *config.MimirConfig
}

func NewMimirTool(config *config.MimirConfig) Mimir {
	return &MimirTool{config: config}
}

func (c *MimirTool) ListRules() (map[string][]models.PrometheusRuleGroup, error) {
	cmd := exec.Command("mimirtool", "rules", "print", "--disable-color")
	cmd.Env = append(cmd.Env, fmt.Sprintf("MIMIR_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("MIMIR_TENANT_ID=%s", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("MIMIR_API_KEY=%s", c.config.ApiKey))
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

func (c *MimirTool) CreateRules(resource models.PrometheusRuleGrouping) (string, error) {
	tmpFile, err := createTmpFile(resource)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cmd := exec.Command("mimirtool", "rules", "load", tmpFile)
	cmd.Env = append(cmd.Env, fmt.Sprintf("MIMIR_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("MIMIR_TENANT_ID=%s", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("MIMIR_API_KEY=%s", c.config.ApiKey))
	res, err := cmd.Output()
	return string(res), err
}
