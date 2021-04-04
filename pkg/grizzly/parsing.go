package grizzly

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
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
	if opts.ResourceFile != nil {
		if strings.HasSuffix(*opts.ResourceFile, ".jsonnet") ||
			strings.HasSuffix(*opts.ResourceFile, ".libsonnet") ||
			strings.HasSuffix(*opts.ResourceFile, ".json") {
			return ParseJsonnet(registry, *opts.ResourceFile, opts)
		}

		if strings.HasSuffix(*opts.ResourceFile, ".yaml") ||
			strings.HasSuffix(*opts.ResourceFile, ".yml") {
			return ParseYAML(registry, *opts.ResourceFile, opts)
		}
	}
	if opts.ConfigFile != nil {
		config, err := ParseConfig(registry, opts)
		if err != nil {
			return nil, err
		}
		var resources Resources
		for _, source := range config.Outbound {
			globs, err := filepath.Glob(source.Path)
			if err != nil {
				return nil, err
			}
			for _, filename := range globs {
				parsedResources, err := ParseYAML(registry, filename, opts)
				if err != nil {
					return nil, err
				}
				resources = append(resources, parsedResources...)
			}
		}
		return resources, nil
	}

	return nil, fmt.Errorf("Either a config file or a resource file is required")
}

// ParseConfig parses a config file from a filename
func ParseConfig(registry Registry, opts GrizzlyOpts) (*Config, error) {
	configResources, err := ParseYAML(registry, *opts.ConfigFile, opts)
	if err != nil {
		return nil, err
	}
	config, err := NewConfig(configResources)
	if err != nil {
		return nil, err
	}
	return config, nil

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
		handler, err := registry.GetParser(m.Kind())
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
func UnparseYAML(resource Resource, source InboundSource) error {
	tmpl, err := template.New("path").Parse(source.Template)
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, resource)
	path := buf.String()
	y, err := resource.YAML()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, []byte(y), 0644)
	if err != nil {
		return err
	}
	return nil
}
