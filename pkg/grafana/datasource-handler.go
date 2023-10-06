package grafana

import (
	"errors"
	"fmt"
	"github.com/grafana/grafana-api-golang-client"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// DatasourceHandler is a grizzly.Handler for Grafana data sources.
type DatasourceHandler struct {
	Provider Provider
}

// NewDatasourceHandler returns a new grizzly.Handler for Grafana data sources.
func NewDatasourceHandler(provider Provider) *DatasourceHandler {
	return &DatasourceHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler.
func (h *DatasourceHandler) Kind() string {
	return "Datasource"
}

// Validate checks that the uid within the resource spec,
// if any, matches with the resource's name, from resource's metadata.
func (h *DatasourceHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource.
func (h *DatasourceHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a data source.
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

// GetUID returns the UID for a resource
func (h *DatasourceHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DatasourceHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return h.getRemoteDatasource(uid)
}

// GetRemote retrieves a datasource as a Resource
func (h *DatasourceHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteDatasource(resource.Name())
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

func (h *DatasourceHandler) getRemoteDatasource(uid string) (*grizzly.Resource, error) {
	ds, err := h.Provider.client.DataSourceByUID(uid)

	var nf gapi.ErrNotFound
	if err != nil && errors.As(err, &nf) {
		ds, err = h.Provider.client.DataSourceByName(uid)
	}

	if err != nil {
		if errors.As(err, &nf) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}

	spec, err := structToMap(ds)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)

	return &resource, nil
}
