package client

import (
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/grizzly/pkg/mimir/models"
)

type MimirFunc func(config.MimirConfig) Mimir

type Mimir interface {
	ListRules() (map[string][]models.PrometheusRuleGroup, error)
	CreateRules(resource models.PrometheusRuleGrouping) error
}
