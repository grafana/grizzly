package grizzly

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// LoggingOpts contains logging options (used in all commands)
type LoggingOpts struct {
	LogLevel string
}

// Opts contains options for most Grizzly commands
type Opts struct {
	LoggingOpts
	Directory    bool
	JsonnetPaths []string
	JsonnetVars  JsonnetVars
	Targets      []string
}

// PreviewOpts contains options to configure a preview
type PreviewOpts struct {
	ExpiresSeconds int
}

type JsonnetVars map[string]string

func (v *JsonnetVars) Type() string {
	return "stringMap"
}

func (v *JsonnetVars) String() string {
	return fmt.Sprintf("%v", *v)
}

func (v *JsonnetVars) Set(val string) error {
	if v == nil {
		return errors.New("nil pointer")
	}

	if *v == nil {
		*v = make(JsonnetVars)
	}

	name, content, err := getVarVal(val)
	if err != nil {
		return err
	}
	(*v)[name] = content

	return nil
}

func getVarVal(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	name := parts[0]
	if len(parts) == 1 {
		content, exists := os.LookupEnv(name)
		if exists {
			return name, content, nil
		}
		return "", "", fmt.Errorf("environment variable %v was undefined", name)
	}
	return name, parts[1], nil
}
