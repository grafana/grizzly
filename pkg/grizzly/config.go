package grizzly

// Config provides configuration to `grizzly`
type Config struct {
	Registry Registry
	Notifier Notifier
}

// GrizzlyOpts contains options for all Grizzly commands
type GrizzlyOpts struct {
	JsonnetPaths []string
	Targets      []string
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
