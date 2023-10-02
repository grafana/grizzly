package encoding

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/goccy/go-yaml"
)

// NewYAMLDecoder returns a YAML decoder configured to unmarshal data from the given reader.
func NewYAMLDecoder(reader io.Reader) *yaml.Decoder {
	return yaml.NewDecoder(reader)
}

// MarshalYAML takes an input and renders as a YAML string.
func MarshalYAML(input any) (string, error) {
	y, err := yaml.MarshalWithOptions(
		input,
		yaml.Indent(4),
		yaml.IndentSequence(true),
		yaml.UseLiteralStyleIfMultiline(true),
		yaml.CustomMarshaler[float64](func(v float64) ([]byte, error) {
			// goccy/go-yaml tends to add .0 suffixes to floats, even when they're not required.
			// To preserve consistency with go-yaml/yaml, this custom marshaler disables that feature.

			if v == math.Inf(0) {
				return []byte(".inf"), nil
			}
			if v == math.Inf(-1) {
				return []byte("-.inf"), nil
			}
			if math.IsNaN(v) {
				return []byte(".nan"), nil
			}

			return []byte(strconv.FormatFloat(v, 'g', -1, 64)), nil
		}),
	)
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
