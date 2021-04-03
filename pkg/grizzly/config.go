package grizzly

// GrizzlyOpts contains options for all Grizzly commands
type GrizzlyOpts struct {
	JsonnetPaths []string
	Targets      []string
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
