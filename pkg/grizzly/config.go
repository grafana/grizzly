package grizzly

// Config provides configuration to `grafana-dash`
type Config struct {
	Registry    Registry
	JsonnetPath string
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
