package grafana

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// DatasourceHandler is a Grizzly Handler for Grafana datasources
type DatasourceHandler struct {
	Provider Provider
}

// NewDatasourceHandler returns a new Grizzly Handler for Grafana datasources
func NewDatasourceHandler(provider Provider) *DatasourceHandler {
	return &DatasourceHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *DatasourceHandler) Kind() string {
	return "Datasource"
}

// APIVersion returns group and version of the provider of this resource
func (h *DatasourceHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a datasource
func (h *DatasourceHandler) GetExtension() string {
	return "json"
}

const (
	datasourceGlob    = "datasources/datasource-*"
	datasourcePattern = "datasources/datasource-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *DatasourceHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, datasourceGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *DatasourceHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(datasourcePattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *DatasourceHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	defaults := map[string]interface{}{
		"basicAuth":         false,
		"basicAuthPassword": "",
		"basicAuthUser":     "",
		"database":          "",
		"orgId":             1,
		"password":          "",
		"secureJsonFields":  map[string]interface{}{},
		"typeLogoUrl":       "",
		"user":              "",
		"withCredentials":   false,
		"readOnly":          false,
	}
	spec := resource["spec"].(map[string]interface{})
	for k := range defaults {
		_, ok := spec[k]
		if !ok {
			spec[k] = defaults[k]
		}
	}
	spec["uid"] = m.Metadata().Name()
	resource["spec"] = spec
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DatasourceHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("version")
	resource.DeleteSpecKey("id")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DatasourceHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	resource.SetSpecValue("id", existing.GetSpecValue("id"))
	resource.DeleteSpecKey("version")
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DatasourceHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return getRemoteDatasource(UID)
}

// GetRemote retrieves a datasource as a Resource
func (h *DatasourceHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return getRemoteDatasource(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *DatasourceHandler) ListRemote() ([]string, error) {
	return getRemoteDatasourceList()
}

// Add pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Add(resource grizzly.Resource) error {
	return postDatasource(resource)
}

// Update pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Update(existing, resource grizzly.Resource) error {
	return putDatasource(resource)
}
