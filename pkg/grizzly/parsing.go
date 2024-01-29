package grizzly

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/grafana/grizzly/pkg/config"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func Parse(resourcePath string, opts Opts) (Resources, error) {
	stat, err := os.Stat(resourcePath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return ParseFile(opts, resourcePath)
	}

	var resources Resources
	files, err := FindResourceFiles(resourcePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		r, err := ParseFile(opts, file)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r...)
	}
	resources = resources.Sort()
	return resources, nil
}

func FindResourceFiles(resourcePath string) ([]string, error) {
	files := []string{}

	err := filepath.Walk(resourcePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func ParseFile(opts Opts, resourceFile string) (Resources, error) {
	switch filepath.Ext(resourceFile) {
	case ".json":
		return ParseJSON(resourceFile, opts)
	case ".yaml", ".yml":
		return ParseYAML(resourceFile, opts)
	case ".jsonnet", ".libsonnet":
		return ParseJsonnet(resourceFile, opts)
	default:
		return nil, fmt.Errorf("%s must be yaml, json or jsonnet", resourceFile)
	}
}

// ParseJSON evaluates a JSON file and parses it into resources
func ParseJSON(resourceFile string, opts Opts) (Resources, error) {

	m := map[string]interface{}{}
	f, err := os.Open(resourceFile)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return nil, err
	}

	var resource Resource
	if opts.OnlySpec {
		resource = newResource(resourceFile, m, opts.ResourceKind, opts.FolderUID)
	} else {
		resource = Resource(m)
	}

	return Resources{resource}, nil
}

// ParseDashboardJSON parses a JSON file with a single dashboard object into a Resources (to align with ParseFile interface)
func newResource(resourceFile string, spec map[string]any, kind, folderUID string) map[string]any {
	uid := strings.ReplaceAll(filepath.Base(resourceFile), filepath.Ext(resourceFile), "")

	handler := Registry.Handlers[kind]

	resource := Resource{
		"apiVersion": handler.APIVersion(),
		"kind":       handler.Kind(),
		"metadata": map[string]interface{}{
			"name": uid,
		},
		"spec": spec,
	}
	if handler.UsesFolders() {
		resource.SetMetadata("folder", folderUID)
	}

	return resource
}

// ParseYAML evaluates a YAML file and parses it into resources
func ParseYAML(yamlFile string, opts Opts) (Resources, error) {
	f, err := os.Open(yamlFile)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	decoder := yaml.NewDecoder(reader)
	manifests := map[string]manifest.Manifest{}
	var m manifest.Manifest
	var resources Resources
	for i := 0; ; i++ {
		err = decoder.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error decoding %s: %v", yamlFile, err)
		}
		manifests[strconv.Itoa(i)] = m
		handler, err := Registry.GetHandler(m.Kind())
		if err != nil {
			return nil, err
		}
		parsedResources, err := handler.Parse(m)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s: %v", yamlFile, err)
		}
		currentContext, err := config.CurrentContext()
		if err != nil {
			return nil, err
		}
		targets := currentContext.GetTargets(opts.Targets)
		for _, parsedResource := range parsedResources {
			if Registry.ResourceMatchesTarget(handler, parsedResource.UID(), targets) {
				resources = append(resources, parsedResource)
			}
		}
	}
	return resources, nil
}

//go:embed grizzly.jsonnet
var script string

// ParseJsonnet evaluates a jsonnet file and parses it into an object tree
func ParseJsonnet(jsonnetFile string, opts Opts) (Resources, error) {

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
		handler, err := Registry.GetHandler(m.Kind())
		if err != nil {
			log.Error("Error getting handler: ", err)
			continue
		}
		parsedResources, err := handler.Parse(m)
		if err != nil {
			return nil, err
		}
		for _, parsedResource := range parsedResources {
			if Registry.ResourceMatchesTarget(handler, parsedResource.UID(), targets) {
				resources = append(resources, parsedResource)
			}
		}
	}
	return resources, nil
}
