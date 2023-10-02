package grizzly

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/grafana/grizzly/pkg/encoding"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	log "github.com/sirupsen/logrus"
)

func Parse(resourcePath string, opts Opts) (Resources, error) {
	if !(opts.Directory) {
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
	switch filepath.Ext(resourceFile) {
	case ".yaml", ".yml":
		return ParseYAML(resourceFile, opts)
	case ".jsonnet", ".libsonnet", ".json":
		return ParseJsonnet(resourceFile, opts)
	default:
		return nil, fmt.Errorf("%s must be yaml, json or jsonnet", resourceFile)
	}
}

// ParseYAML evaluates a YAML file and parses it into resources
func ParseYAML(yamlFile string, opts Opts) (Resources, error) {
	f, err := os.Open(yamlFile)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	decoder := encoding.NewYAMLDecoder(reader)
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

// ParseJsonnet evaluates a jsonnet file and parses it into an object tree
func ParseJsonnet(jsonnetFile string, opts Opts) (Resources, error) {
	extracted, err := encoding.ParseJsonnet(jsonnetFile, opts.JsonnetPaths)
	if err != nil {
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
