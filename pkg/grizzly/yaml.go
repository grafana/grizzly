package grizzly

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type YAMLParser struct {
	registry Registry
	logger   *log.Entry
}

func NewYAMLParser(registry Registry) *YAMLParser {
	return &YAMLParser{
		registry: registry,
		logger:   log.WithField("parser", "yaml"),
	}
}

func (parser *YAMLParser) Accept(file string) bool {
	extension := filepath.Ext(file)

	return extension == ".yaml" || extension == ".yml"
}

// Parse evaluates a YAML file and parses it into resources
func (parser *YAMLParser) Parse(file string, options ParserOptions) (Resources, error) {
	parser.logger.WithField("file", file).Debug("Parsing file")

	f, err := os.Open(file)
	if err != nil {
		return Resources{}, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	decoder := yaml.NewDecoder(reader)
	resources := NewResources()
	for i := 0; ; i++ {
		var m any
		err = decoder.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Resources{}, err
		}

		source := Source{
			Format:     formatYAML,
			Path:       file,
			Rewritable: true,
		}
		parsedResources, err := parseAny(parser.registry, m, options.DefaultResourceKind, options.DefaultFolderUID, source)
		if err != nil {
			return Resources{}, err
		}

		resources.Merge(parsedResources)
	}

	return resources, nil
}
