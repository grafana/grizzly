package grizzly

import "github.com/grafana/tanka/pkg/kubernetes/manifest"

// NewManifest returns a new Tanka Manifest based off a Kubernetes style object
func NewManifest(handler Handler, name string, spec interface{}) (manifest.Manifest, error) {
	m := map[string]interface{}{}
	m["apiVersion"] = handler.APIVersion()
	m["kind"] = handler.Kind()
	m["spec"] = spec
	metadata := map[string]interface{}{}
	metadata["name"] = name
	m["metadata"] = metadata
	return manifest.New(m)
}
