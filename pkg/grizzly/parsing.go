package grizzly

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/google/go-jsonnet"
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

	return resources, nil
}

func FindResourceFiles(resourcePath string) ([]string, error) {
	var files []string
	for _, handler := range Registry.Handlers {
		handlerFiles, err := handler.FindResourceFiles(resourcePath)
		if err != nil {
			return nil, err
		}
		files = append(files, handlerFiles...)
	}
	return files, nil
}

func ParseFile(opts Opts, resourceFile string) (Resources, error) {
	if opts.JSONSpec && filepath.Ext(resourceFile) != ".json" {
		return nil, fmt.Errorf("when -s flag is passed, command expects only json files as resources")
	}

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

func manifestFile(resourceFile string) (bool, error) {
	if filepath.Ext(resourceFile) != ".json" {
		return false, nil
	}

	m := map[string]interface{}{}

	f, err := os.Open(resourceFile)
	if err != nil {
		return false, err
	}

	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return false, err
	}

	if _, ok := m["spec"]; ok {
		return true, nil
	}

	return false, nil
}

// ParseJSON evaluates a JSON file and parses it into resources
func ParseJSON(resourceFile string, opts Opts) (Resources, error) {
	if opts.JSONSpec {
		return ParseDashboardJSON(resourceFile, opts)
	}

	isManifest, err := manifestFile(resourceFile)
	if err != nil {
		return Resources{}, err
	}

	// TODO: refactor, no need to read the file twice
	if !isManifest {
		return ParseDashboardJSON(resourceFile, opts)
	}

	return ParseJsonnet(resourceFile, opts)
}

// ParseDashboardJSON parses a JSON file with a single dashboard object into a Resources (to align with ParseFile interface)
func ParseDashboardJSON(jsonFile string, opts Opts) (Resources, error) {
	if filepath.Ext(jsonFile) != ".json" {
		return nil, fmt.Errorf("when -s flag is passed, command expects only json files as resources")
	}

	f, err := os.Open(jsonFile)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	err = json.NewDecoder(f).Decode(&spec)
	if err != nil {
		return Resources{}, err
	}

	handler := Registry.Handlers["Dashboard"]

	resource := Resource{
		"apiVersion": handler.APIVersion(),
		"kind":       handler.Kind(),
		"metadata": map[string]interface{}{
			"folder": opts.FolderUID,
			"name":   spec["uid"],
		},
		"spec": spec,
	}

	return Resources{resource}, nil
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
			return nil, err
		}
		log.Printf("Decoding %d", i)
		manifests[strconv.Itoa(i)] = m
		handler, err := Registry.GetHandler(m.Kind())
		if err != nil {
			return nil, err
		}
		parsedResources, err := handler.Parse(m)
		if err != nil {
			return nil, err
		}
		for _, parsedResource := range parsedResources {
			if parsedResource.MatchesTarget(opts.Targets) {
				resources = append(resources, parsedResource)
			}
		}
	}
	sort.Sort(resources)
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
			if parsedResource.MatchesTarget(opts.Targets) {
				resources = append(resources, parsedResource)
			}
		}
	}
	sort.Sort(resources)
	return resources, nil
}

// MarshalYAML takes a resource and renders it to a source file as a YAML string
func MarshalYAML(resource Resource, filename string) error {
	y, err := resource.YAML()
	if err != nil {
		return err
	}
	return writeFile(filename, []byte(y))
}

func MarshalSpecToJSON(resource Resource, filename string) error {
	j, err := json.MarshalIndent(resource.Spec(), "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filename, j)
}

func writeFile(filename string, content []byte) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
