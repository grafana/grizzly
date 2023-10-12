package grizzly

// LoggingOpts contains logging options (used in all commands)
type LoggingOpts struct {
	LogLevel string
}

// Opts contains options for most Grizzly commands
type Opts struct {
	LoggingOpts
	Directory    bool
	JsonnetPaths []string
	Targets      []string

	// Used for supporting commands that output dashboard JSON
	FolderUID string
	JSONSpec  bool
}

// PreviewOpts contains options to configure a preview
type PreviewOpts struct {
	ExpiresSeconds int
}
