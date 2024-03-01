package config

type GrafanaConfig struct {
	URL                string `yaml:"url" mapstructure:"url"`
	User               string `yaml:"user" mapstructure:"user"`
	Token              string `yaml:"token" mapstructure:"token"`
	InsecureSkipVerify bool   `yaml:"insecure-skip-verify" mapstructure:"insecure-skip-verify"`
}

type MimirConfig struct {
	Address  string `yaml:"address" mapstructure:"address"`
	TenantID int64  `yaml:"tenant-id" mapstructure:"tenant-id"`
	ApiKey   string `yaml:"api-key" mapstructure:"api-key"`
}

type SyntheticMonitoringConfig struct {
	Token     string `yaml:"token" mapstructure:"token"`
	StackID   int64  `yaml:"stack-id" mapstructure:"stack-id"`
	LogsID    int64  `yaml:"logs-id" mapstructure:"logs-id"`
	MetricsID int64  `yaml:"metrics-id" mapstructure:"metrics-id"`
}

type Context struct {
	Name                string                    `yaml:"name" mapstructure:"name"`
	Grafana             GrafanaConfig             `yaml:"grafana" mapstructure:"grafana"`
	Mimir               MimirConfig               `yaml:"mimir" mapstructure:"mimir"`
	SyntheticMonitoring SyntheticMonitoringConfig `yaml:"synthetic-monitoring" mapstructure:"synthetic-monitoring"`
	Targets             []string                  `yaml:"targets" mapstructure:"targets"`
	OutputFormat        string                    `yaml:"output-format" mapstructure:"output-format"`
	OnlySpec            bool                      `yaml:"only-spec" mapstructure:"only-spec"`
	ResourceKind        string                    `yaml:"resource-kind" mapstructure:"resource-kind"`
	FolderUID           string                    `yaml:"folder-uid" mapstructure:"folder-uid"`
}
