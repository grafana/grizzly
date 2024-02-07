package grizzly

// LoggingOpts contains logging options (used in all commands)
type LoggingOpts struct {
	LogLevel string
}

// Opts contains options for most Grizzly commands
type Opts struct {
	LoggingOpts
	Directory    bool // Deprecated: now is gathered with os.Stat(<resource-path>)
	JsonnetPaths []string
	Targets      []string
	OutputFormat string

	// Used for supporting resources without envelopes
	OnlySpec     bool
	HasOnlySpec  bool
	FolderUID    string
	ResourceKind string
}

// PreviewOpts contains options to configure a preview
type PreviewOpts struct {
	ExpiresSeconds int
}
