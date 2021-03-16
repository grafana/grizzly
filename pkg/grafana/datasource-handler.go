package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/mitchellh/mapstructure"
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

func (h *DatasourceHandler) newDatasourceResource(path, uid, filename string, source Datasource) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Filename: filename,
		Handler:  h,
		Detail:   source,
		JSONPath: path,
	}
	return resource
}

// Parse parses a manifest object into a struct for this resource type
func (h *DatasourceHandler) Parse(m manifest.Manifest) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	spec := m["spec"].(map[string]interface{})
	source := Datasource{
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
	err := mapstructure.Decode(spec, &source)
	if err != nil {
		return nil, err
	}
	source["name"] = m.Metadata().Name()
	resource := h.newDatasourceResource("", source.UID(), "", source)
	key := resource.Key()
	resources[key] = resource

	return resources, nil
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
	resource := h.newDatasourceResource(datasourcesPath, UID, "", *source)
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
	resource := h.newDatasourceResource(datasourcesPath, uid, "", *source)
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

func (h *DatasourceHandler) delete(resource grizzly.Resource, key string) {
	delete(resource.Detail.(Datasource), key)
}
