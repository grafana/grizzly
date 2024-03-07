package grizzly

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type YAMLParser struct {
	registry Registry
}

func NewYAMLParser(registry Registry) *YAMLParser {
	return &YAMLParser{
		registry: registry,
	}
}

func (parser *YAMLParser) Accept(file string) bool {
	extension := filepath.Ext(file)

	return extension == ".yaml" || extension == ".yml"
}

// Parse evaluates a YAML file and parses it into resources
func (parser *YAMLParser) Parse(file string, options ParserOptions) (Resources, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	decoder := yaml.NewDecoder(reader)
	var resources Resources
	for i := 0; ; i++ {
		var m map[string]any
		err = decoder.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding %s: %v", file, err)
		}

		parsedResources, err := parseAny(parser.registry, m, options.DefaultResourceKind, options.DefaultFolderUID)
		if err != nil {
			return nil, err
		}

		resources = append(resources, parsedResources...)
	}

	return resources, nil
}
