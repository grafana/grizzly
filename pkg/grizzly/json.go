package grizzly

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type JSONParser struct {
	registry Registry
}

func NewJSONParser(registry Registry) *JSONParser {
	return &JSONParser{
		registry: registry,
	}
}

func (parser *JSONParser) Accept(file string) bool {
	return filepath.Ext(file) == ".json"
}

// Parse evaluates a JSON file and parses it into resources
func (parser *JSONParser) Parse(file string, options ParserOptions) (Resources, error) {
	f, err := os.Open(file)
	if err != nil {
		return Resources{}, err
	}
	defer f.Close()

	m := map[string]any{}
	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return Resources{}, err
	}

	source := Source{
		Format:     "json",
		Path:       file,
		Rewritable: true,
	}

	return parseAny(parser.registry, m, options.DefaultResourceKind, options.DefaultFolderUID, source)
}
