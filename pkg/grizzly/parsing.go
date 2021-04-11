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
	if opts.ResourceFile != nil && opts.Directory != nil {
		return nil, fmt.Errorf("Only specify a directory or a resource file, not both")
	}
	if opts.ResourceFile != nil {
		return ParseFile(registry, opts, *opts.ResourceFile)
	}
	if opts.Directory != nil {
		resources := Resources{}
		files, err := FindResourceFiles(registry, opts, *opts.Directory)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			log.Println("Parsing", file)
			r, err := ParseFile(registry, opts, file)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r...)
		}
		return resources, nil
	}
	return nil, fmt.Errorf("Must specify either a resource file or a directory")
}

func FindResourceFiles(registry Registry, opts GrizzlyOpts, dir string) ([]string, error) {
	files := []string{}
	for _, handler := range registry.Handlers {
		handlerFiles, err := handler.FindResourceFiles(dir)
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
