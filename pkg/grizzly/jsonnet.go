package grizzly

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
)

// ExtendedImporter does stuff
type ExtendedImporter struct {
	loaders    []importLoader    // for loading jsonnet from somewhere. First one that returns non-nil is used
	processors []importProcessor // for post-processing (e.g. yaml -> json)
}

type importLoader func(importedFrom, importedPath string) (c *jsonnet.Contents, foundAt string, err error)

// importProcessor are executed after the file import and may modify the result
// further
type importProcessor func(contents, foundAt string) (c *jsonnet.Contents, err error)

// newFileLoader returns an importLoader that uses jsonnet.FileImporter to source
// files from the local filesystem
func newFileLoader(fi *jsonnet.FileImporter) importLoader {
	return func(importedFrom, importedPath string) (contents *jsonnet.Contents, foundAt string, err error) {
		var c jsonnet.Contents
		c, foundAt, err = fi.Import(importedFrom, importedPath)
		return &c, foundAt, err
	}
}

func newExtendedImporter(jsonnetFile, path string, jpath []string) *ExtendedImporter {
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
	return &ExtendedImporter{
		loaders: []importLoader{
			newFileLoader(&jsonnet.FileImporter{
				JPaths: absolutePaths,
			})},
		processors: []importProcessor{},
	}
}

// Import implements the functionality offered by the ExtendedImporter
func (i *ExtendedImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
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

func GenerateJsonnetImports(dir, out string) error {
	var b strings.Builder
	b.WriteString("{\n")
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			ext := filepath.Ext(d.Name())
			base := strings.TrimSuffix(d.Name(), ext)
			b.WriteString(fmt.Sprintf("  \"%s\": {\n", base))
			b.WriteString(fmt.Sprintf("    \"path\": \"%s\",\n", path))
			b.WriteString(fmt.Sprintf("    \"type\": \"%s\",\n", strings.TrimPrefix(ext, ".")))
			if ext == ".yaml" || ext == ".yml" {
				b.WriteString(fmt.Sprintf("    \"resource\": std.parseYAML(importstr \"%s\"),\n", path))
			} else if ext == ".json" {
				b.WriteString(fmt.Sprintf("    \"resource\": import \"%s\",\n", path))
			}
			b.WriteString("  },\n")
		}
		return nil
	})
	if err != nil {
		return err
	}
	b.WriteString("}\n")
	return os.WriteFile(out, []byte(b.String()), 0644)
}
