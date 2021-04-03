package grizzly

// GrizzlyOpts contains options for all Grizzly commands
type GrizzlyOpts struct {
	ConfigFile   *string
	ResourceFile *string
	JsonnetPaths *[]string
	Targets      *[]string
}

func (o *GrizzlyOpts) ConsumeArguments(args []string) {
	if len(args) > 0 {
		o.ResourceFile = &args[0]
	}
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
