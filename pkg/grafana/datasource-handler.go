package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/mitchellh/mapstructure"
)

// DatasourceHandler is a Grizzly Provider for Grafana datasources
type DatasourceHandler struct{}

// NewDatasourceHandler returns configuration defining a new Grafana Provider
func NewDatasourceHandler() *DatasourceHandler {
	return &DatasourceHandler{}
}

// GetName returns the name for this provider
func (h *DatasourceHandler) GetName() string {
	return "datasource"
}

// GetFullName returns the name for this provider
func (h *DatasourceHandler) GetFullName() string {
	return "grafana.datasource"
}

const datasourcesPath = "grafanaDatasources"

// GetJSONPaths returns paths within Jsonnet output that this provider will consume
func (h *DatasourceHandler) GetJSONPaths() []string {
	return []string{
		datasourcesPath,
	}
}

// GetExtension returns the file name extension for a datasource
func (h *DatasourceHandler) GetExtension() string {
	return "json"
}

// APIVersion returns the api version for this resource
func (h *DatasourceHandler) APIVersion() string {
	return "grafana.com/v1"
}

// Kind returns the resource kind for this type of resource
func (h *DatasourceHandler) Kind() string {
	return "Datasource"
}
func (h *DatasourceHandler) newDatasourceResource(uid, filename string, source Datasource) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Filename: filename,
		Handler:  h,
		Detail:   source,
		JSONPath: datasourcesPath,
	}
	return resource
}

// ParseHiddenElements parses an interface{} object into a struct for this resource type
func (h *DatasourceHandler) ParseHiddenElements(path string, i interface{}) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		m, err := grizzly.NewManifest(h, k, v)
		if err != nil {
			return nil, err
		}
		resource, err := h.Parse(m)
		if err != nil {
			return nil, err
		}
		resources[resource.Key()] = *resource
	}
	return resources, nil
}

// Parse parses a single resource from an interface{} object
func (h *DatasourceHandler) Parse(m manifest.Manifest) (*grizzly.Resource, error) {
	source := Datasource{}
	source["basicAuth"] = false
	source["basicAuthPassword"] = ""
	source["basicAuthUser"] = ""
	source["database"] = ""
	source["orgId"] = 1
	source["password"] = ""
	source["secureJsonFields"] = map[string]interface{}{}
	source["typeLogoUrl"] = ""
	source["user"] = ""
	source["withCredentials"] = false
	source["readOnly"] = false

	err := mapstructure.Decode(m["spec"], &source)
	if err != nil {
		return nil, err
	}
	resource := h.newDatasourceResource(source.UID(), m.Metadata().Name(), source)
	return &resource, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DatasourceHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	h.delete(resource, "version")
	h.delete(resource, "id")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DatasourceHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	resource.Detail.(Datasource)["id"] = existing.Detail.(Datasource)["id"]
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DatasourceHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	source, err := getRemoteDatasource(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving datasource %s: %v", UID, err)
	}
	resource := h.newDatasourceResource(UID, "", *source)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (h *DatasourceHandler) GetRepresentation(uid string, resource grizzly.Resource) (string, error) {
	j, err := json.MarshalIndent(resource.Detail, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// GetRemoteRepresentation retrieves a datasource as JSON
func (h *DatasourceHandler) GetRemoteRepresentation(uid string) (string, error) {
	source, err := getRemoteDatasource(uid)
	if err != nil {
		return "", err
	}
	return source.toJSON()
}

// GetRemote retrieves a datasource as a Resource
func (h *DatasourceHandler) GetRemote(uid string) (*grizzly.Resource, error) {
	source, err := getRemoteDatasource(uid)
	if err != nil {
		return nil, err
	}
	resource := h.newDatasourceResource(uid, "", *source)
	return &resource, nil
}

// Add pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Add(resource grizzly.Resource) error {
	return postDatasource(newDatasource(resource))
}

// Update pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Update(existing, resource grizzly.Resource) error {
	return putDatasource(newDatasource(resource))
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DatasourceHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
	return grizzly.ErrNotImplemented
}

func (h *DatasourceHandler) delete(resource grizzly.Resource, key string) {
	delete(resource.Detail.(Datasource), key)
}
