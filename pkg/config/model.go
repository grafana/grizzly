package config

type GrafanaConfig struct {
	URL   string `yaml:"url"`
	User  string `yaml:"user"`
	Token string `yaml:"token"`
}

type MimirConfig struct {
	Address  string `yaml:"address"`
	TenantID int64  `yaml:"tenant-id"`
	ApiKey   string `yaml:"api-key"`
}

type SyntheticMonitoringConfig struct {
	Token     string `yaml:"token"`
	StackID   int64  `yaml:"stack-id"`
	LogsID    int64  `yaml:"logs-id"`
	MetricsID int64  `yaml:"metrics-id"`
}

type Context struct {
	Name                string                    `yaml:"name"`
	Grafana             GrafanaConfig             `yaml:"grafana"`
	Mimir               MimirConfig               `yaml:"mimir"`
	SyntheticMonitoring SyntheticMonitoringConfig `yaml:"synthetic-monitoring"`
	Targets             []string                  `yaml:"targets"`
}
