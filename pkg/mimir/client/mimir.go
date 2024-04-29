package client

import (
	"github.com/grafana/grizzly/pkg/mimir/models"
)

type Mimir interface {
	ListRules() (map[string][]models.PrometheusRuleGroup, error)
	CreateRules(resource models.PrometheusRuleGrouping) error
}
