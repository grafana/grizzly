package models

// PrometheusRuleGroup encapsulates a list of rules
type PrometheusRuleGroup struct {
	Namespace string                   `yaml:"-"`
	Name      string                   `yaml:"name"`
	Rules     []map[string]interface{} `yaml:"rules"`
}

// PrometheusRuleGrouping encapsulates a set of named rule groups
type PrometheusRuleGrouping struct {
	Namespace string                `json:"namespace"`
	Groups    []PrometheusRuleGroup `json:"groups"`
}
