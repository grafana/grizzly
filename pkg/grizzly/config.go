package grizzly

// Opts contains options for all Grizzly commands
type Opts struct {
	Directory    bool
	JsonnetPaths []string
	Targets      []string
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
