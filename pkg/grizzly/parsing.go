package grizzly

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ParseError struct {
	File string
	Err  error
}

func (err ParseError) Error() string {
	return fmt.Sprintf("parse error in '%s': %s", err.File, err.Err)
}

type ParserOptions struct {
	DefaultResourceKind string
	DefaultFolderUID    string
}

type FormatParser interface {
	Accept(file string) bool
	Parse(file string, options ParserOptions) (Resources, error)
}

type Parser interface {
	Parse(resourcePath string, options ParserOptions) (Resources, error)
}

type parsersConfig struct {
	continueOnError bool
}

type ParserOpt func(config *parsersConfig)

func ParserContinueOnError(continueOnError bool) ParserOpt {
	return func(config *parsersConfig) {
		config.continueOnError = continueOnError
	}
}

func DefaultParser(registry Registry, targets []string, jsonnetPaths []string, opts ...ParserOpt) Parser {
	config := &parsersConfig{}

	for _, opt := range opts {
		opt(config)
	}

	return NewFilteredParser(
		registry,
		NewChainParser([]FormatParser{
			NewJSONParser(registry),
			NewYAMLParser(registry),
			NewJsonnetParser(registry, jsonnetPaths),
		}, config.continueOnError),
		targets,
	)
}

type FilteredParser struct {
	registry  Registry
	decorated Parser
	targets   []string
}

func NewFilteredParser(registry Registry, decorated Parser, targets []string) *FilteredParser {
	return &FilteredParser{
		registry:  registry,
		decorated: decorated,
		targets:   targets,
	}
}

func (parser *FilteredParser) Parse(resourcePath string, options ParserOptions) (Resources, error) {
	resources, err := parser.decorated.Parse(resourcePath, options)
	if err != nil {
		return resources, err
	}

	resources = resources.Filter(func(resource Resource) bool {
		return parser.registry.ResourceMatchesTarget(resource.Kind(), resource.Name(), parser.targets)
	})

	return parser.registry.Sort(resources), nil
}

type ChainParser struct {
	formatParsers   []FormatParser
	continueOnError bool
}

func NewChainParser(formatParsers []FormatParser, continueOnError bool) *ChainParser {
	return &ChainParser{
		formatParsers:   formatParsers,
		continueOnError: continueOnError,
	}
}

func (parser *ChainParser) Parse(resourcePath string, options ParserOptions) (Resources, error) {
	if resourcePath == "" {
		return nil, nil
	}

	stat, err := os.Stat(resourcePath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return parser.parseFile(resourcePath, options)
	}

	var parsedResources Resources
	var finalErr error
	_ = filepath.WalkDir(resourcePath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		r, err := parser.parseFile(path, options)
		if err != nil {
			finalErr = multierror.Append(finalErr, err)

			if !parser.continueOnError {
				return err
			} else {
				return nil
			}
		}

		parsedResources = append(parsedResources, r...)

		return nil
	})

	return parsedResources, finalErr
}

func (parser *ChainParser) parseFile(file string, options ParserOptions) (Resources, error) {
	for _, l := range parser.formatParsers {
		if !l.Accept(file) {
			continue
		}

		resources, err := l.Parse(file, options)
		if err != nil {
			return nil, ParseError{File: file, Err: err}
		}
		return resources, nil
	}

	return nil, fmt.Errorf("unrecognized format for %s", file)
}

func parseAny(registry Registry, data any, resourceKind, folderUID string, source Source) (Resources, error) {
	hasEnvelope := DetectEnvelope(data)
	if hasEnvelope {
		m := data.(map[string]any)
		err := ValidateEnvelope(m)
		if err != nil {
			return nil, err
		}

		handler, err := registry.GetHandler(m["kind"].(string))
		if err != nil {
			return nil, err
		}

		resource, err := handler.Parse(m)
		if err != nil {
			return nil, err
		}
		resource.SetSource(source)
		return Resources{*resource}, nil
	}

	kind := registry.Detect(data)
	if kind == "" && resourceKind != "" {
		kind = resourceKind
	}

	if kind != "" {
		handler, err := registry.GetHandler(kind)
		if err != nil {
			return nil, err
		}

		if handler.UsesFolders() && folderUID == "" {
			// TODO: the error shouldn't assume a CLI environment
			return nil, fmt.Errorf("folder (-f) required with --onlyspec")
		}

		m := data.(map[string]any)
		resource, err := NewResource(handler.APIVersion(), handler.Kind(), "dummy", m)
		if err != nil {
			return nil, err
		}
		resource.SetSource(source)

		uid, err := handler.GetSpecUID(resource)
		if err != nil {
			return nil, err
		}

		resource.SetMetadata("name", uid)
		if handler.UsesFolders() {
			resource.SetMetadata("folder", folderUID)
		}

		r, err := handler.Parse(resource.Body)
		if err != nil {
			return nil, err
		}

		return Resources{*r}, nil
	}

	walker := walker{
		source: source,
	}
	err := walker.Walk(data)

	return walker.Resources, err
}

// DetectEnvelope identifies whether this resource is enveloped or not
func DetectEnvelope(data any) bool {
	m, ok := data.(map[string]any)
	if !ok {
		return false
	}
	expectedKeys := []string{
		"kind",
		"metadata",
		"spec",
	}
	for _, key := range expectedKeys {
		_, ok := m[key]
		if !ok {
			return false
		}
	}
	return true
}

// ValidateEnvelope confirms that this resource is a complete enveloped resource
func ValidateEnvelope(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("resource is not a map")
	}
	errors := []string{}

	kind, ok := m["kind"]
	if !ok || kind == "" {
		errors = append(errors, "kind missing")
	}
	metadata, ok := m["metadata"]
	if !ok {
		errors = append(errors, "metadata missing")
	} else {
		m, ok := metadata.(map[string]interface{})
		if !ok {
			errors = append(errors, "Metadata is not a map")
		} else {
			name, ok := m["name"]
			if !ok || name == nil || name == "" {
				errors = append(errors, "metadata/name missing")
			} else {
				n, ok := name.(string)
				if !ok {
					errors = append(errors, "metadata/name is not a string")
				}
				if n == "" {
					errors = append(errors, "metadata/name is blank")
				}
			}
		}
	}
	spec, ok := m["spec"]
	if !ok || spec == nil {
		errors = append(errors, "spec missing")
	} else {
		s, ok := spec.(map[string]any)
		if !ok {
			errors = append(errors, "spec is not a map")
		} else {
			if len(s) == 0 {
				errors = append(errors, "spec should not be empty")
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors parsing resource: %s", strings.Join(errors, ", "))
	}
	return nil
}

type walker struct {
	Resources Resources
	source    Source
}

// Walk scans the raw interface{} for objects that look like enveloped objects and
// extracts those into a list
func (w *walker) Walk(raw interface{}) error {
	if err := w.walkJSON(raw, nil); err != nil {
		return err
	}
	return nil
}

// walkJSON recurses into either a map or list, returning a list of all objects that look
// like enveloped resources. We support resources at an arbitrary level of nesting, and
// return an error if a node is not walkable.
//
// Handling the different types is quite gross, so we split this method into a generic
// walkJSON, and then walkObj/walkList to handle the two different types of collection we
// support.
func (w *walker) walkJSON(raw interface{}, path trace) error {
	// check for known types
	switch v := raw.(type) {
	case map[string]interface{}:
		return w.walkObj(v, path)
	case []interface{}:
		return w.walkList(v, path)
	}

	log.Debugf("recursion ended on key %q of type %T which does not belong to a valid resource", path.Name(), raw)
	return ErrorPrimitiveReached{
		path:      path.Base(),
		key:       path.Name(),
		primitive: raw,
	}
}

func (w *walker) walkList(list []interface{}, path trace) error {
	for idx, value := range list {
		err := w.walkJSON(value, append(path, fmt.Sprintf("[%d]", idx)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *walker) walkObj(obj map[string]any, path trace) error {
	validateErr := ValidateEnvelope(obj)
	if validateErr != nil {
		// this is not an envelope, skip.
	} else {
		resource, err := ResourceFromMap(obj)
		if err != nil {
			return err
		}

		source := w.source
		source.Location = path.Full()
		resource.SetSource(source)
		w.Resources = append(w.Resources, *resource)
		return nil
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		path := append(path, key)
		if obj[key] == nil { // result from false if condition in Jsonnet
			continue
		}
		err := w.walkJSON(obj[key], path)
		if err != nil {
			return err.(ErrorPrimitiveReached).WithContainingObj(obj, validateErr)
		}
	}

	return nil
}

type trace []string

func (t trace) Full() string {
	return "." + strings.Join(t, ".")
}

func (t trace) Base() string {
	if len(t) > 0 {
		t = t[:len(t)-1]
	}
	return "." + strings.Join(t, ".")
}

func (t trace) Name() string {
	if len(t) > 0 {
		return t[len(t)-1]
	}

	return ""
}

// ErrorPrimitiveReached occurs when walkJSON reaches the end of nested dicts without finding a valid resource
type ErrorPrimitiveReached struct {
	path, key        string
	primitive        interface{}
	containingObj    map[string]any
	containingObjErr error
}

func (e ErrorPrimitiveReached) WithContainingObj(obj map[string]any, err error) ErrorPrimitiveReached {
	if e.containingObj == nil {
		e.containingObj = obj
		e.containingObjErr = err
	}
	return e
}

func (e ErrorPrimitiveReached) Error() string {
	errMessage := fmt.Sprintf(`found invalid object (at %s): %s`, e.path, e.containingObjErr)

	container, err := yaml.Marshal(e.containingObj)
	if err != nil {
		log.Errorf("failed to marshal invalid object: %s", err)
	} else {
		errMessage += "\n\n" + string(container)
	}

	return errMessage
}
