package manifests

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"gopkg.in/yaml.v2"
)

// New creates a new manifest from a string
func New(kind, name string, data, spec interface{}) (*manifest.Manifest, error) {
	m := manifest.Manifest{}
	m["apiVersion"] = "grizzly.grafana.com/v1alpha1"
	m["kind"] = kind
	m["metadata"] = map[string]interface{}{
		"name": name,
	}
	m, err := ParseData(m, data)

	if err != nil {
		return nil, err
	}
	if spec != nil {
		switch s := spec.(type) {
		case map[string]interface{}:
			m["spec"] = s
		case []byte:
			msi := map[string]interface{}{}
			if err := json.Unmarshal(s, &msi); err != nil {
				return nil, fmt.Errorf("Error: %v\n%s", err, string(s))
			}
			m["spec"] = msi
		}
	}
	return &m, nil
}

// JSON converts a manifest into a json string
func JSON(m manifest.Manifest) (string, error) {
	j, err := json.MarshalIndent(m, "  ", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// YAML converts a manifest into a yaml string
func YAML(m manifest.Manifest) (string, error) {
	j, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// SpecAsJSON converts a manifest into a json string
func SpecAsJSON(m manifest.Manifest) (string, error) {
	j, err := json.MarshalIndent(m["spec"], "  ", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// SpecAsYAML converts a manifest into a yaml string
func SpecAsYAML(m manifest.Manifest) (string, error) {
	j, err := yaml.Marshal(m["spec"])
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// DataAsJSON converts a manifest into a json string
func DataAsJSON(m manifest.Manifest) (string, error) {
	j, err := json.MarshalIndent(m["data"], "  ", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// ParseData parses data element as JSON
func ParseData(m manifest.Manifest, data interface{}) (manifest.Manifest, error) {
	if data != nil {
		switch d := data.(type) {
		case map[string]interface{}:
			m["data"] = d
		case string:
			msi := map[string]interface{}{}
			if err := json.Unmarshal([]byte(d), &msi); err != nil {
				return nil, fmt.Errorf("Error: %v\n%s", err, string(d))
			}
			m["data"] = msi
		case []byte:
			msi := map[string]interface{}{}
			if err := json.Unmarshal(d, &msi); err != nil {
				return nil, fmt.Errorf("Error: %v\n%s", err, string(d))
			}
			m["data"] = msi
		default:
			return nil, fmt.Errorf("Can't parse %T into manifest.data", data)
		}
	}
	return m, nil
}
