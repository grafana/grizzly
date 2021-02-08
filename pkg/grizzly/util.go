package grizzly

import "github.com/grafana/tanka/pkg/kubernetes/manifest"

// NewManifest returns a new Tanka Manifest based off a Kubernetes style object
func NewManifest(APIVersion, kind, name string, spec interface{}) manifest.Manifest {
	m := manifest.Manifest{
		"apiVersion": APIVersion,
		"kind":       kind,
		"spec":       spec,
		"metadata": map[string]interface{}{
			name: name,
		},
	}
	return m
}
