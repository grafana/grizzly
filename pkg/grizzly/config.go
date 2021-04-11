package grizzly

// GrizzlyOpts contains options for all Grizzly commands
type GrizzlyOpts struct {
	Directory    *string
	ResourceFile *string
	JsonnetPaths *[]string
	Targets      *[]string
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
