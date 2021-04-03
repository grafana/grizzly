package grizzly

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
	"gopkg.in/yaml.v3"
)

func Parse(registry Registry, opts GrizzlyOpts) (Resources, error) {
	if strings.HasSuffix(*opts.ResourceFile, ".yaml") ||
		strings.HasSuffix(*opts.ResourceFile, ".yml") {
		return ParseYAML(registry, *opts.ResourceFile, opts)
	} else if strings.HasSuffix(*opts.ResourceFile, ".jsonnet") ||
		strings.HasSuffix(*opts.ResourceFile, ".libsonnet") ||
		strings.HasSuffix(*opts.ResourceFile, ".json") {
		return ParseJsonnet(registry, *opts.ResourceFile, opts)
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
