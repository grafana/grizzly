package grizzly

// Config provides configuration to `grizzly`
type Config struct {
	Registry    Registry
	Notifier    Notifier
	JsonnetPath string
}

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
