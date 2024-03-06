package grizzly

import (
	_ "embed" // used to embed grizzly.jsonnet script below
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

type JsonnerParser struct {
	registry     Registry
	jsonnetPaths []string
}

func NewJsonnerParser(registry Registry, jsonnetPaths []string) *JsonnerParser {
	return &JsonnerParser{
		registry:     registry,
		jsonnetPaths: jsonnetPaths,
	}
}

func (parser *JsonnerParser) Accept(file string) bool {
	extension := filepath.Ext(file)

	return extension == ".jsonnet" || extension == ".libsonnet"
}

// Parse evaluates a jsonnet file and parses it into an object tree
func (parser *JsonnerParser) Parse(file string, options Options) (Resources, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", file)
	}
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	result, err := evaluateJsonnet(file, currentWorkingDirectory, parser.jsonnetPaths)
	if err != nil {
		return nil, err
	}
	var data interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	resources, err := parseAny(parser.registry, data, options.DefaultResourceKind, options.DefaultFolderUID)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// extendedImporter does stuff
type extendedImporter struct {
	loaders    []importLoader    // for loading jsonnet from somewhere. First one that returns non-nil is used
	processors []importProcessor // for post-processing (e.g. yaml -> json)
}

type importLoader func(importedFrom, importedPath string) (c *jsonnet.Contents, foundAt string, err error)

// importProcessor are executed after the file import and may modify the result
// further
type importProcessor func(contents, foundAt string) (c *jsonnet.Contents, err error)

//go:embed grizzly.jsonnet
var script string

func evaluateJsonnet(jsonnetFile, wd string, jpath []string) (string, error) {
	s := fmt.Sprintf(script, jsonnetFile)
	vm := jsonnet.MakeVM()
	vm.Importer(newExtendedImporter(jsonnetFile, wd, jpath))
	vm.NativeFunction(escapeStringRegexNativeFunc())
	vm.NativeFunction(regexMatchNativeFunc())
	vm.NativeFunction(regexSubstNativeFunc())

	return vm.EvaluateAnonymousSnippet(jsonnetFile, s)
}

// newFileLoader returns an importLoader that uses jsonnet.FileImporter to source
// files from the local filesystem
func newFileLoader(fi *jsonnet.FileImporter) importLoader {
	return func(importedFrom, importedPath string) (contents *jsonnet.Contents, foundAt string, err error) {
		var c jsonnet.Contents
		c, foundAt, err = fi.Import(importedFrom, importedPath)
		return &c, foundAt, err
	}
}

func newExtendedImporter(jsonnetFile, path string, jpath []string) *extendedImporter {
	absolutePaths := make([]string, len(jpath)*2+1)
	absolutePaths = append(absolutePaths, path)
	jsonnetDir := filepath.Dir(jsonnetFile)
	for _, p := range jpath {
		if !filepath.IsAbs(p) {
			p = filepath.Join(jsonnetDir, p)
		}
		absolutePaths = append(absolutePaths, p)
	}
	for _, p := range jpath {
		if !filepath.IsAbs(p) {
			p = filepath.Join(path, p)
		}
		absolutePaths = append(absolutePaths, p)
	}
	return &extendedImporter{
		loaders: []importLoader{
			newFileLoader(&jsonnet.FileImporter{
				JPaths: absolutePaths,
			})},
		processors: []importProcessor{},
	}
}

// Import implements the functionality offered by the extendedImporter
func (i *extendedImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	// load using loader
	for _, loader := range i.loaders {
		c, f, err := loader(importedFrom, importedPath)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}
		if c != nil {
			contents = *c
			foundAt = f
			break
		}
	}

	// check if needs postprocessing
	for _, processor := range i.processors {
		c, err := processor(contents.String(), foundAt)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}
		if c != nil {
			contents = *c
			break
		}
	}

	return contents, foundAt, nil
}

// escapeStringRegexNativeFunc escapes all regular expression metacharacters
// and returns a regular expression that matches the literal text.
func escapeStringRegexNativeFunc() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "escapeStringRegex",
		Params: ast.Identifiers{"str"},
		Func: func(s []interface{}) (interface{}, error) {
			return regexp.QuoteMeta(s[0].(string)), nil
		},
	}
}

// regexMatchNativeFunc returns whether the given string is matched by the given re2 regular expression.
func regexMatchNativeFunc() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "regexMatch",
		Params: ast.Identifiers{"regex", "string"},
		Func: func(s []interface{}) (interface{}, error) {
			return regexp.MatchString(s[0].(string), s[1].(string))
		},
	}
}

// regexSubstNativeFunc replaces all matches of the re2 regular expression with another string.
func regexSubstNativeFunc() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "regexSubst",
		Params: ast.Identifiers{"regex", "src", "repl"},
		Func: func(data []interface{}) (interface{}, error) {
			regex, src, repl := data[0].(string), data[1].(string), data[2].(string)

			r, err := regexp.Compile(regex)
			if err != nil {
				return "", err
			}
			return r.ReplaceAllString(src, repl), nil
		},
	}
}
