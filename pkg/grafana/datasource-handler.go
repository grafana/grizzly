package grafana

import (
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

func (h *DatasourceHandler) newDatasourceResource(m manifest.Manifest) grizzly.Resource {
	resource := grizzly.Resource{
		UID:     m.Metadata().Name(),
		Detail:  m,
		Handler: h,
	}
	return resource
}

// GetRemoteByUID retrieves a dashboard as a resource
func (h *DatasourceHandler) GetRemoteByUID(uid string) (*grizzly.Resource, error) {
	m, err := getRemoteDatasource(uid)
	if err != nil {
		return nil, err
	}
	return grizzly.NewResource(*m, h), nil
}

// GetRemote retrieves a dashboard as a resource
func (h *DatasourceHandler) GetRemote(existing grizzly.Resource) (*grizzly.Resource, error) {
	return h.GetRemoteByUID(existing.Detail.Metadata().Name())
}

// Add pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Add(resource grizzly.Resource) error {
	return postDatasource(resource.Detail)
}

// Update pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Update(existing, resource grizzly.Resource) error {
	return putDatasource(resource.Detail)
}
