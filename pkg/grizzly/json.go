package grizzly

import (
	"encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type JSONParser struct {
	registry Registry
	logger   *log.Entry
}

func NewJSONParser(registry Registry) *JSONParser {
	return &JSONParser{
		registry: registry,
		logger:   log.WithField("parser", "json"),
	}
}

func (parser *JSONParser) Accept(file string) bool {
	return filepath.Ext(file) == ".json"
}

// Parse evaluates a JSON file and parses it into resources
func (parser *JSONParser) Parse(file string, options ParserOptions) (Resources, error) {
	parser.logger.WithField("file", file).Debug("Parsing file")

	f, err := os.Open(file)
	if err != nil {
		return Resources{}, err
	}
	defer f.Close()

	var m any
	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return Resources{}, err
	}

	source := Source{
		Format:     formatJSON,
		Path:       file,
		Rewritable: true,
	}

	return parseAny(parser.registry, m, options.DefaultResourceKind, options.DefaultFolderUID, source)
}
