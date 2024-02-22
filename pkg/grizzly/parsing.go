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
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func Parse(registry Registry, resourcePath string, opts *Opts) (Resources, error) {
	stat, err := os.Stat(resourcePath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return ParseFile(registry, *opts, resourcePath)
	}
	opts.IsDir = true

	var resources Resources
	files, err := FindResourceFiles(resourcePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		r, err := ParseFile(registry, *opts, file)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r...)
	}
	resources = registry.Sort(resources)
	return resources, nil
}

func FindResourceFiles(resourcePath string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(resourcePath, func(path string, info fs.DirEntry, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func ParseFile(registry Registry, opts Opts, resourceFile string) (Resources, error) {
	switch filepath.Ext(resourceFile) {
	case ".json":
		return ParseJSON(registry, resourceFile, opts)
	case ".yaml", ".yml":
		return ParseYAML(registry, resourceFile, opts)
	case ".jsonnet", ".libsonnet":
		return ParseJsonnet(registry, resourceFile, opts)
	default:
		return nil, fmt.Errorf("%s must be yaml, json or jsonnet", resourceFile)
	}
}

// ParseJSON evaluates a JSON file and parses it into resources
func ParseJSON(registry Registry, resourceFile string, opts Opts) (Resources, error) {
	f, err := os.Open(resourceFile)
	if err != nil {
		return nil, err
	}

	m := map[string]any{}
	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return nil, err
	}
	hasEnvelope := DetectEnvelope(m)
	if !hasEnvelope {
		kind := registry.Detect(m)
		if kind == "" {
			if opts.ResourceKind == "" {
				return nil, fmt.Errorf("cannot deduce kind of %s", resourceFile)
			}
			kind = opts.ResourceKind
		}
		resources, err := newOnlySpecResources(registry, m, kind, opts.FolderUID)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", resourceFile, err)
		}
		return resources, nil

	} else {
		resources, err := newWithEnvelopeResources(registry, m)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", resourceFile, err)
		}
		return resources, nil
	}
}

// ParseYAML evaluates a YAML file and parses it into resources
func ParseYAML(registry Registry, yamlFile string, opts Opts) (Resources, error) {
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
			return nil, fmt.Errorf("xError decoding %s: %v", yamlFile, err)
		}
		var parsedResources Resources

		var kind string
		hasEnvelope := DetectEnvelope(m)
		if !hasEnvelope {
			kind = registry.Detect(m)
			if kind == "" {
				if opts.ResourceKind == "" {
					return nil, fmt.Errorf("cannot deduce kind of %s", yamlFile)
				}
				kind = opts.ResourceKind
			}
			parsedResources, err = newOnlySpecResources(registry, m, kind, opts.FolderUID)
			if err != nil {
				return nil, fmt.Errorf("Error parsing %s: %v", yamlFile, err)
			}
		} else {
			parsedResources, err = newWithEnvelopeResources(registry, m)
			if err != nil {
				return nil, fmt.Errorf("Error parsing %s: %v", yamlFile, err)
			}
			kind = parsedResources[0].Kind()
		}
		handler, err := registry.GetHandler(kind)
		if err != nil {
			return nil, err
		}
		currentContext, err := config.CurrentContext()
		if err != nil {
			return nil, err
		}
		targets := currentContext.GetTargets(opts.Targets)
		for _, parsedResource := range parsedResources {
			if registry.ResourceMatchesTarget(handler, parsedResource.Name(), targets) {
				resources = append(resources, parsedResource)
			}
		}
	}
	return resources, nil
}

func newOnlySpecResources(registry Registry, data map[string]any, kind, folderUID string) (Resources, error) {
	if kind == "" {
		return nil, fmt.Errorf("Kind (-k) required with --onlyspec")
	}
	handler, err := registry.GetHandler(kind)
	if err != nil {
		return nil, err
	}
	if handler.UsesFolders() && folderUID == "" {
		return nil, fmt.Errorf("folder (-f) required with --onlyspec")
	}
	resource, err := NewResource(handler.APIVersion(), handler.Kind(), "dummy", data)
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
	m := manifest.Manifest(resource)
	return handler.Parse(m)
}

func newWithEnvelopeResources(registry Registry, data map[string]any) (Resources, error) {
	err := ValidateEnvelope(data)
	if err != nil {
		return nil, err
	}

	resource, err := ResourceFromMap(data)
	if err != nil {
		return nil, err
	}
	handler, err := registry.GetHandler(resource.Kind())
	if err != nil {
		return nil, err
	}
	m := manifest.Manifest(resource)
	return handler.Parse(m)
}

// DetectEnvelope identifies whether this resource is enveloped or not
func DetectEnvelope(data map[string]any) bool {
	expectedKeys := []string{
		"kind",
		"metadata",
		"spec",
	}
	for _, key := range expectedKeys {
		_, ok := data[key]
		if !ok {
			return false
		}
	}
	return true
}

// ValidateEnvelope confirms that this resource is a complete enveloped resource
func ValidateEnvelope(data map[string]any) error {
	errors := []string{}

	kind, ok := data["kind"]
	if !ok || kind == "" {
		errors = append(errors, "kind missing")
	}
	metadata, ok := data["metadata"]
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
	spec, ok := data["spec"]
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

//go:embed grizzly.jsonnet
var script string

// ParseJsonnet evaluates a jsonnet file and parses it into an object tree
func ParseJsonnet(registry Registry, jsonnetFile string, opts Opts) (Resources, error) {

	if _, err := os.Stat(jsonnetFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", jsonnetFile)
	}
	script := fmt.Sprintf(script, jsonnetFile)
	vm := jsonnet.MakeVM()
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	vm.Importer(newExtendedImporter(jsonnetFile, currentWorkingDirectory, opts.JsonnetPaths))
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

	extracted, err := process.Extract(data)
	if err != nil {
		return nil, err
	}

	// Unwrap *List types
	if err := process.Unwrap(extracted); err != nil {
		return nil, err
	}
	currentContext, err := config.CurrentContext()
	if err != nil {
		return nil, err
	}
	targets := currentContext.GetTargets(opts.Targets)
	resources := Resources{}
	for _, m := range extracted {
		handler, err := registry.GetHandler(m.Kind())
		if err != nil {
			log.Error("Error getting handler: ", err)
			continue
		}
		parsedResources, err := handler.Parse(m)
		if err != nil {
			return nil, err
		}
		for _, parsedResource := range parsedResources {
			err := ValidateEnvelope(parsedResource)
			if err != nil {
				return nil, err
			}

			if registry.ResourceMatchesTarget(handler, parsedResource.Name(), targets) {
				resources = append(resources, parsedResource)
			}
		}
	}
	return resources, nil
}
