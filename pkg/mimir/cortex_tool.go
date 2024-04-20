package mimir

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/grafana/grizzly/pkg/config"
)

type CortexTool interface {
	ExecuteCortexTool(args ...string) ([]byte, error)
}

type Cortex struct {
	config *config.MimirConfig
}

func NewCortexTool(config *config.MimirConfig) *Cortex {
	return &Cortex{config: config}
}

func (c *Cortex) ExecuteCortexTool(args ...string) ([]byte, error) {
	path := os.Getenv("CORTEXTOOL_PATH")
	if path == "" {
		var err error
		path, err = exec.LookPath("cortextool")
		if err != nil {
			return nil, err
		} else if path == "" {
			return nil, fmt.Errorf("cortextool not found")
		}
	}
	cmd := exec.Command(path, args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_ADDRESS=%s", c.config.Address))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_TENANT_ID=%s", c.config.TenantID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CORTEX_API_KEY=%s", c.config.APIKey))
	return exec.Command(path, args...).Output()
}
