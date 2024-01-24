package grizzly

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"strings"
)

//nolint:gochecknoglobals
var templates *template.Template

//go:embed embed/templates/*/*.tmpl embed/assets/*
//nolint:gochecknoglobals
var embedFS embed.FS

//nolint:gochecknoinits
func init() {
	base := template.New("grizzly")
	base.Option("missingkey=error")

	templates = template.Must(findAndParseTemplates(embedFS, base, "embed/templates"))
}

func findAndParseTemplates(vfs fs.FS, rootTmpl *template.Template, rootDir string) (*template.Template, error) {
	err := fs.WalkDir(vfs, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileHandle, err := vfs.Open(path)
		if err != nil {
			return err
		}

		contents, err := io.ReadAll(fileHandle)
		if err != nil {
			return err
		}

		templateName := strings.TrimPrefix(strings.TrimPrefix(path, rootDir), "/")
		t := rootTmpl.New(templateName)
		_, err = t.Parse(string(contents))

		return err
	})

	return rootTmpl, err
}
