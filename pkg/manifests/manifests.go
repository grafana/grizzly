package manifests

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

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
		default:
			log.Printf("Unknown type %T", spec)
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

// JoinUID joins elements into a multipart UID
func JoinUID(elements ...string) string {
	return strings.Join(elements, ".")
}

// SplitUID splits a multipart UID into its parts
func SplitUID(uid string) []string {
	isDelimiter := func(c rune) bool {
		return c == '/' || c == '.'
	}
	return strings.FieldsFunc(uid, isDelimiter)

}

// GetMetadata gets a value from the metadata of a manifest
func GetMetadata(m *manifest.Manifest, key string) string {
	metadata := (*m)["metadata"].(map[string]interface{})
	return metadata[key].(string)
}

// SetMetadata sets a value in the metadata of a manifest
func SetMetadata(m *manifest.Manifest, key, value string) *manifest.Manifest {
	metadata := (*m)["metadata"].(map[string]interface{})
	metadata[key] = value
	(*m)["metadata"] = metadata
	return m
}

// RemoveSpecFields removes a set of field values from the spec within the manifest
func RemoveSpecFields(m *manifest.Manifest, keys []string) *manifest.Manifest {
	spec := (*m)["spec"].(map[string]interface{})
	for _, key := range keys {
		delete(spec, key)
	}
	(*m)["spec"] = spec
	return m
}

// GetSpecField gets a spec field value from within the manifest
func GetSpecField(m *manifest.Manifest, key string) interface{} {
	spec := (*m)["spec"].(map[string]interface{})
	return spec[key]
}

// SetSpecField sets a spec field value within the manifest
func SetSpecField(m *manifest.Manifest, key string, value interface{}) *manifest.Manifest {
	spec := (*m)["spec"].(map[string]interface{})
	spec[key] = value
	(*m)["spec"] = spec
	return m
}
