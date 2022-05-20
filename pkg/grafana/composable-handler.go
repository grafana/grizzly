package grafana

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// ComposableHandler is a Grizzly Handler for Grafana datasources
type ComposableHandler struct {
	Provider     Provider
	ResourceKind grizzly.ResourceKind
}

// NewComposableHandler returns a new Grizzly Handler for Grafana datasources
func NewComposableHandler(provider Provider, resourceKind grizzly.ResourceKind) *ComposableHandler {
	return &ComposableHandler{
		Provider:     provider,
		ResourceKind: resourceKind,
	}
}

// Kind returns the kind for this handler
func (h *ComposableHandler) Kind() string {
	return h.ResourceKind.Kind
}

func (h *ComposableHandler) Compose(resource grizzly.Resource, context grizzly.Resources) (grizzly.Resources, error) {
	resolver := grizzly.NewResolver(context, h.ResourceKind, h.Provider.ResourceKinds())
	return resolver.Resolve(resource)
}

// Validate returns the uid of resource
func (h *ComposableHandler) Validate(resource grizzly.Resource) error {
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *ComposableHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a datasource
func (h *ComposableHandler) GetExtension() string {
	return "json"
}

// FindResourceFiles identifies files within a directory that this handler can process
func (h *ComposableHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, "composable") // @TODO composableGlob
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *ComposableHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(datasourcePattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *ComposableHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *ComposableHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *ComposableHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *ComposableHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *ComposableHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return nil, grizzly.ErrIsComposable
}

// GetRemote retrieves a datasource as a Resource
func (h *ComposableHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return nil, grizzly.ErrIsComposable
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *ComposableHandler) ListRemote() ([]string, error) {
	return nil, grizzly.ErrIsComposable
}

// Add pushes a datasource to Grafana via the API
func (h *ComposableHandler) Add(resource grizzly.Resource) error {
	return grizzly.ErrIsComposable
}

// Update pushes a datasource to Grafana via the API
func (h *ComposableHandler) Update(existing, resource grizzly.Resource) error {
	return grizzly.ErrIsComposable
}
