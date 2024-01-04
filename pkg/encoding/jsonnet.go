package encoding

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-jsonnet"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
)

//go:embed grizzly.jsonnet
var script string

// ParseJsonnet evaluates a jsonnet file and parses it into an object tree
func ParseJsonnet(jsonnetFile string, jsonnetPaths []string) (map[string]manifest.Manifest, error) {
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

	result, err := vm.EvaluateAnonymousSnippet(jsonnetFile, script)
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

	return extracted, nil
}
