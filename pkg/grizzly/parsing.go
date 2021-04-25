package grizzly

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
	"gopkg.in/yaml.v3"
)

func Parse(registry Registry, opts GrizzlyOpts) (Resources, error) {
	if !(*opts.Directory) {
		return ParseFile(registry, opts, *opts.ResourcePath)
	}
	resources := Resources{}
	files, err := FindResourceFiles(registry, opts)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		r, err := ParseFile(registry, opts, file)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r...)
	}
	return resources, nil
}

func FindResourceFiles(registry Registry, opts GrizzlyOpts) ([]string, error) {
	files := []string{}
	for _, handler := range registry.Handlers {
		handlerFiles, err := handler.FindResourceFiles(*opts.ResourcePath)
		if err != nil {
			return nil, err
		}
		files = append(files, handlerFiles...)
	}
	return files, nil
}

func ParseFile(registry Registry, opts GrizzlyOpts, resourceFile string) (Resources, error) {
	if strings.HasSuffix(resourceFile, ".yaml") ||
		strings.HasSuffix(resourceFile, ".yml") {
		return ParseYAML(registry, resourceFile, opts)
	} else if strings.HasSuffix(resourceFile, ".jsonnet") ||
		strings.HasSuffix(resourceFile, ".libsonnet") ||
		strings.HasSuffix(resourceFile, ".json") {
		return ParseJsonnet(registry, resourceFile, opts)
	} else {
		return nil, fmt.Errorf("Either a config file or a resource file is required")
	}
}

// ParseYAML evaluates a YAML file and parses it into resources
func ParseYAML(registry Registry, yamlFile string, opts GrizzlyOpts) (Resources, error) {
	f, err := os.Open(yamlFile)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	decoder := yaml.NewDecoder(reader)
	manifests := map[string]manifest.Manifest{}
	var m manifest.Manifest
	var resources Resources
	for i := 0; decoder.Decode(&m) == nil; i++ {
		manifests[strconv.Itoa(i)] = m
		handler, err := registry.GetHandler(m.Kind())
		if err != nil {
			return nil, err
		}
		parsedResources, err := handler.Parse(m)
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
func ParseJsonnet(registry Registry, jsonnetFile string, opts GrizzlyOpts) (Resources, error) {

	script := fmt.Sprintf(script, jsonnetFile)
	vm := jsonnet.MakeVM()
	vm.Importer(newExtendedImporter(*opts.JsonnetPaths))
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
		handler, err := registry.GetHandler(m.Kind())
		if err != nil {
			log.Println("Error getting handler", err)
			continue
		}
		parsedResources, err := handler.Parse(m)
		if err != nil {
			return nil, err
		}
		resources = append(resources, parsedResources...)
	}
	return resources, nil
}

// UnparseYAML takes a resource and renders it to a source file as a YAML string
func UnparseYAML(resource Resource, filename string) error {
	y, err := resource.YAML()
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, []byte(y), 0644)
	if err != nil {
		return err
	}
	return nil
}
