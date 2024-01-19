package grizzly

import (
	"os"
	"path/filepath"
)

func Format(resourcePath string, resource *Resource, format string, onlySpec bool) ([]byte, string, string, error) {
	var content string
	var filename string
	var extension string
	var err error

	spec := resource
	if onlySpec {
		s := Resource(resource.Spec())
		spec = &s
	}

	switch format {
	case "yaml":
		extension = "yaml"
		filename, err = getFilename(resourcePath, resource, extension)
		if err != nil {
			return nil, "", "", err
		}
		content, err = spec.YAML()
	case "json":
		extension = "json"
		filename, err = getFilename(resourcePath, resource, extension)
		if err != nil {
			return nil, "", "", err
		}
		content, err = spec.JSON()
	default:
		extension = "yaml"
		filename, err = getFilename(resourcePath, resource, extension)
		if err != nil {
			return nil, "", "", err
		}
		content, err = spec.YAML()
	}
	return []byte(content), filename, extension, err
}

func getFilename(resourcePath string, resource *Resource, extension string) (string, error) {
	handler, err := Registry.GetHandler(resource.Kind())
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
