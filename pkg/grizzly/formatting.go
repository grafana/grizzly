package grizzly

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	formatJSON    = "json"
	formatYAML    = "yaml"
	formatWide    = "wide"
	formatDefault = "default"
)

func Format(registry Registry, resourcePath string, resource *Resource, format string, onlySpec bool) ([]byte, string, string, error) {
	var content []byte
	var filename string
	var extension string
	var err error

	spec := resource.Body
	if onlySpec {
		spec = resource.Spec()
	}

	if format == formatJSON {
		extension = formatJSON
		j, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return nil, "", "", err
		}
		content = j
	} else {
		extension = formatYAML
		y, err := yaml.Marshal(spec)
		if err != nil {
			return nil, "", "", err
		}
		content = y
	}
	filename, err = getFilename(registry, resourcePath, resource, extension)
	if err != nil {
		return nil, "", "", err
	}

	return content, filename, extension, nil
}

func getFilename(registry Registry, resourcePath string, resource *Resource, extension string) (string, error) {
	handler, err := registry.GetHandler(resource.Kind())
	if err != nil {
		return "", err
	}
	return filepath.Join(resourcePath, handler.ResourceFilePath(*resource, extension)), nil
}

func WriteFile(filename string, content []byte) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
