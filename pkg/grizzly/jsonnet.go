package grizzly

import (
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

func newExtendedImporter(jpath []string) *ExtendedImporter {
	return &ExtendedImporter{
		loaders: []importLoader{
			newFileLoader(&jsonnet.FileImporter{
				JPaths: jpath,
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
