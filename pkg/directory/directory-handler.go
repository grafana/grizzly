package directory

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// DirectoryHandler is a Grizzly Handler for Grafana datasources
type DirectoryHandler struct {
	Provider Provider
}

// NewDirectoryHandler returns a new Grizzly Handler for Grafana datasources
func NewDirectoryHandler(provider Provider) *DirectoryHandler {
	return &DirectoryHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *DirectoryHandler) Kind() string {
	return "Datasource"
}

// APIVersion returns group and version of the provider of this resource
func (h *DirectoryHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetJSONPaths returns paths within Jsonnet output that this provider will consume
func (h *DirectoryHandler) GetJSONPaths() []string {
	return []string{}
}

// GetExtension returns the file name extension for a datasource
func (h *DirectoryHandler) GetExtension() string {
	return "json"
}

// Parse parses a manifest object into a struct for this resource type
func (h *DirectoryHandler) Parse(source string, m manifest.Manifest) (grizzly.ResourceList, error) {
	resource := grizzly.Resource(m)
	path := resource.GetSpecString("path")
	if resource.HasSpecString("glob") {

	} else {
		dir := filepath.Dir(source)
		fullpath := filepath.Join(dir, path)
		files, err := ioutil.ReadDir(fullpath)
		if err != nil {
			return nil, nil
		}
		for _, file := range files {

		}

	}
	return resource.AsResourceList(), nil
}

type ErrNotSupported struct {
	Action string
}

func (e ErrNotSupported) Error() string {
	return fmt.Sprintf("%s not supported by DirectoryHandler", e.Action)
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DirectoryHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DirectoryHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetByUID does nothing
func (h *DirectoryHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return nil, ErrNotSupported{}
}

// GetRemote does nothing
func (h *DirectoryHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return nil, ErrNotSupported{}
}

// Add does nothing
func (h *DirectoryHandler) Add(resource grizzly.Resource) error {
	return ErrNotSupported{}
}

// Update does nothing
func (h *DirectoryHandler) Update(existing, resource grizzly.Resource) error {
	return ErrNotSupported{}
}
