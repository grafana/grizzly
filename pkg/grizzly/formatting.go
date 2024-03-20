package grizzly

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Format(registry Registry, resourcePath string, resource *Resource, format string, onlySpec bool) ([]byte, string, string, error) {
	var content string
	var filename string
	var extension string
	var err error

	spec := resource.Body
	if onlySpec {
		spec = resource.Spec()
	}

	if format == "json" {
		extension = "json"
		j, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return nil, "", "", err
		}
		content = string(j)
	} else {
		extension = "yaml"
		y, err := yaml.Marshal(spec)
		if err != nil {
			return nil, "", "", err
		}
		content = string(y)
	}
	filename, err = getFilename(registry, resourcePath, resource, extension)
	if err != nil {
		return nil, "", "", err
	}

	return []byte(content), filename, extension, nil
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
