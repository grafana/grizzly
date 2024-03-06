package grizzly

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

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

func DefaultParser(registry Registry, targets []string, jsonnetPaths []string) Parser {
	return NewFilteredParser(
		registry,
		NewChainParser([]FormatParser{
			NewJSONParser(registry),
			NewYAMLParser(registry),
			NewJsonnetParser(registry, jsonnetPaths),
		}),
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
		return nil, err
	}

	resources = resources.Filter(func(resource Resource) bool {
		return parser.registry.ResourceMatchesTarget(resource.Kind(), resource.Name(), parser.targets)
	})

	return parser.registry.Sort(resources), nil
}

type ChainParser struct {
	formatParsers []FormatParser
}

func NewChainParser(formatParsers []FormatParser) *ChainParser {
	return &ChainParser{
		formatParsers: formatParsers,
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
	err = filepath.WalkDir(resourcePath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		r, err := parser.parseFile(path, options)
		if err != nil {
			return err
		}

		parsedResources = append(parsedResources, r...)

		return nil
	})

	return parsedResources, err
}

func (parser *ChainParser) parseFile(file string, options ParserOptions) (Resources, error) {
	for _, l := range parser.formatParsers {
		if l.Accept(file) {
			return l.Parse(file, options)
		}
	}

	return nil, fmt.Errorf("unrecognized format for %s", file)
}

func parseAny(registry Registry, data any, resourceKind, folderUID string) (Resources, error) {
	hasEnvelope := DetectEnvelope(data)
	if hasEnvelope {
		m := data.(map[string]any)
		err := ValidateEnvelope(m)
		if err != nil {
			return nil, err
		}

		resource, err := ResourceFromMap(m)
		if err != nil {
			return nil, err
		}
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return nil, err
		}
		return handler.Parse(resource)
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
		uid, err := handler.GetSpecUID(resource)
		if err != nil {
			return nil, err
		}
		resource.SetMetadata("name", uid)

		if handler.UsesFolders() {
			resource.SetMetadata("folder", folderUID)
		}
		return handler.Parse(resource)
	}
	walker := walker{}
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

		w.Resources = append(w.Resources, resource)
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
