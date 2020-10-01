package grizzly

import (
	"github.com/fatih/color"
)

// Config provides configuration to `grafana-dash`
type Config struct {
	Registry    Registry
	JsonnetPath string
}

var (
	// Red makes terminal output red
	Red = color.New(color.FgRed).SprintFunc()
	// Yellow makes terminal output red
	Yellow = color.New(color.FgYellow).SprintFunc()
	// Green makes terminal output red
	Green = color.New(color.FgGreen).SprintFunc()
)

// PreviewOpts Options to Configure a Preview
type PreviewOpts struct {
	ExpiresSeconds int
}
