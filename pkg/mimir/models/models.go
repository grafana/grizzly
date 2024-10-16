package models

// PrometheusRuleGroup encapsulates a list of rules
type PrometheusRuleGroup struct {
	Name  string        `yaml:"name"`
	Rules []interface{} `yaml:"rules"`
}

// PrometheusRuleGrouping encapsulates a set of named rule groups
type PrometheusRuleGrouping struct {
	Namespace string                `yaml:"namespace"`
	Groups    []PrometheusRuleGroup `yaml:"groups"`
}

type PrometheusAlertmanagerConfig struct {
	TemplateFiles      map[string]string `yaml:"template_files"`
	AlertmanagerConfig string            `yaml:"alertmanager_config"`
}
