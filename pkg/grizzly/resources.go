package grizzly

import (
	"strings"

	"gopkg.in/yaml.v2"
)

// Resource is a general representation of something Grizzly manages. It might
// be a Grafana Dashboard, a Prometheus rule or something entirely different.
// Resources are processed by providers. If no provider processes the resource,
// it is ignored.
type Resource interface {
	// Kind should return the type of the resource, e.g. `Dashboard`. Kind must
	// be unique across grizzly
	Kind() string

	// UID must uniquely identify a resource withing its kind. It should be
	// human readable. `Name` was not chosen because many resource
	// impelmentations already have `Name` struct keys.
	UID() string
}

type Resources []Resource

// String prints all resources as a YAML stream. Resource kind and name are
// embedded as comments.
func (res Resources) String() string {
	var b strings.Builder
	for _, r := range res {
		data, err := yaml.Marshal(r)
		if err != nil {
			panic(err)
		}
		b.WriteString("---\n")
		b.WriteString("# kind: " + r.Kind() + "\n")
		b.WriteString("# name: " + r.UID() + "\n")
		b.Write(data)
	}

	return b.String()
}
