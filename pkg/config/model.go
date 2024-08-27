package config

type GrafanaConfig struct {
	URL                string `yaml:"url" mapstructure:"url"`
	User               string `yaml:"user" mapstructure:"user"`
	Token              string `yaml:"token" mapstructure:"token"`
	InsecureSkipVerify bool   `yaml:"insecure-skip-verify" mapstructure:"insecure-skip-verify"`
	TLSHost            string `yaml:"tls-host" mapstructure:"tls-host"`
}

type MimirConfig struct {
	Address  string         `yaml:"address" mapstructure:"address"`
	TenantID string         `yaml:"tenant-id" mapstructure:"tenant-id"`
	APIKey   string         `yaml:"api-key" mapstructure:"api-key"`
	TLS      MimirTLSConfig `yaml:"tls" mapstructure:"tls"`
}

type MimirTLSConfig struct {
	ClientCertPath string `yaml:"client-cert-path,omitempty" mapstructure:"client-cert-path"`
	ClientKeyPath  string `yaml:"client-key-path,omitempty" mapstructure:"client-key-path"`
	CAPath         string `yaml:"ca-path" mapstructure:"ca-path"`
}

type SyntheticMonitoringConfig struct {
	URL string `yaml:"url" mapstructure:"url"`
	// SM can be configured with a metrics publisher token (and various stack information) or an access token gotten from the UI
	Token       string `yaml:"token" mapstructure:"token"`
	StackID     int64  `yaml:"stack-id" mapstructure:"stack-id"`
	LogsID      int64  `yaml:"logs-id" mapstructure:"logs-id"`
	MetricsID   int64  `yaml:"metrics-id" mapstructure:"metrics-id"`
	AccessToken string `yaml:"access-token" mapstructure:"access-token"`
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
