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

	// Used for supporting commands that output dashboard JSON
	FolderUID   string
	OnlySpec    bool
	HasOnlySpec bool
}

// PreviewOpts contains options to configure a preview
type PreviewOpts struct {
	ExpiresSeconds int
}
