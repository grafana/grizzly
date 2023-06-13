package encoding

import (
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// NewYAMLDecoder returns a YAML decoder configured to unmarshal data from the given reader.
func NewYAMLDecoder(reader io.Reader) *yaml.Decoder {
	return yaml.NewDecoder(reader)
}

// MarshalYAML takes an input and renders as a YAML string.
func MarshalYAML(input any) (string, error) {
	y, err := yaml.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(y), nil
}

// MarshalYAMLFile takes an input and renders it to a file as a YAML string.
func MarshalYAMLFile(input any, filename string) error {
	y, err := MarshalYAML(input)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, []byte(y), 0644)
	if err != nil {
		return err
	}

	return nil
}

// UnmarshalYAML takes YAML content as input unmarshals it into the destination.
func UnmarshalYAML(input []byte, destination any) error {
	return yaml.Unmarshal(input, destination)
}
