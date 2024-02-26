package grizzly

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func Parse(registry Registry, resourcePath, resourceKind, folderUID string, targets, jsonnetPaths []string) (Resources, error) {
	stat, err := os.Stat(resourcePath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return ParseFile(registry, resourcePath, resourceKind, folderUID, jsonnetPaths)
	}

	var files []string
	_ = filepath.WalkDir(resourcePath, func(path string, info fs.DirEntry, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	var parsedResources Resources
	for _, file := range files {
		r, err := ParseFile(registry, file, resourceKind, folderUID, jsonnetPaths)
		if err != nil {
			return nil, err
		}
		parsedResources = append(parsedResources, r...)
	}
	var resources Resources
	for _, parsedResource := range parsedResources {
		if registry.ResourceMatchesTarget(parsedResource.Kind(), parsedResource.Name(), targets) {
			resources = append(resources, parsedResource)
		}
	}
	resources = registry.Sort(resources)
	return resources, nil
}

func ParseFile(registry Registry, resourceFile, resourceKind, folderUID string, jsonnetPaths []string) (Resources, error) {
	switch filepath.Ext(resourceFile) {
	case ".json":
		return ParseJSON(registry, resourceFile, resourceKind, folderUID)
	case ".yaml", ".yml":
		return ParseYAML(registry, resourceFile, resourceKind, folderUID)
	case ".jsonnet", ".libsonnet":
		return ParseJsonnet(registry, resourceFile, jsonnetPaths, resourceKind, folderUID)
	default:
		return nil, fmt.Errorf("%s must be yaml, json or jsonnet", resourceFile)
	}
}

// ParseJSON evaluates a JSON file and parses it into resources
func ParseJSON(registry Registry, resourceFile, resourceKind, folderUID string) (Resources, error) {
	f, err := os.Open(resourceFile)
	if err != nil {
		return nil, err
	}

	m := map[string]any{}
	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return nil, err
	}

	resources, err := parseAny(registry, m, resourceKind, folderUID)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", resourceFile, err)
	}
	return resources, err
}

// ParseYAML evaluates a YAML file and parses it into resources
func ParseYAML(registry Registry, yamlFile, resourceKind, folderUID string) (Resources, error) {
	f, err := os.Open(yamlFile)
	if err != nil {
		return nil, err
	}
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
			return nil, fmt.Errorf("Error decoding %s: %v", yamlFile, err)
		}
		parsedResources, err := parseAny(registry, m, resourceKind, folderUID)
		if err != nil {
			return nil, err
		}
		resources = append(resources, parsedResources...)
	}
	return resources, nil
}

//go:embed grizzly.jsonnet
var script string

// ParseJsonnet evaluates a jsonnet file and parses it into an object tree
func ParseJsonnet(registry Registry, jsonnetFile string, jsonnetPaths []string, resourceKind, folderUID string) (Resources, error) {

	if _, err := os.Stat(jsonnetFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", jsonnetFile)
	}
	script := fmt.Sprintf(script, jsonnetFile)
	vm := jsonnet.MakeVM()
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	vm.Importer(newExtendedImporter(jsonnetFile, currentWorkingDirectory, jsonnetPaths))
	for _, nf := range native.Funcs() {
		vm.NativeFunction(nf)
	}

	result, err := vm.EvaluateSnippet(jsonnetFile, script)
	if err != nil {
		return nil, err
	}
	var data interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	resources, err := parseAny(registry, data, resourceKind, folderUID)
	if err != nil {
		return nil, err
	}
	return resources, nil
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
