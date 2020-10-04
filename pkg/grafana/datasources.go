package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/grafana/grizzly/pkg/grizzly"
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

// GetJSONPath returns a paths within Jsonnet output that this provider will consume
func (h *DatasourceHandler) GetJSONPath() string {
	return "grafanaDatasources"
}

// GetExtension returns the file name extension for a datasource
func (h *DatasourceHandler) GetExtension() string {
	return "json"
}

func (h *DatasourceHandler) newDatasourceResource(uid, filename string, source Datasource) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Filename: filename,
		Handler:  h,
		Detail:   source,
		Path:     h.GetJSONPath(),
	}
	return resource
}

// Parse parses an interface{} object into a struct for this resource type
func (h *DatasourceHandler) Parse(i interface{}) (grizzly.Resources, error) {
	resources := grizzly.Resources{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		source := Datasource{}
		err := mapstructure.Decode(v, &source)
		if err != nil {
			return nil, err
		}
		resource := h.newDatasourceResource(source.UID(), k, source)
		key := resource.Key()
		resources[key] = resource
	}
	return resources, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DatasourceHandler) Unprepare(detail map[string]interface{}) map[string]interface{} {
	delete(detail, "version")
	delete(detail, "id")
	return detail
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DatasourceHandler) Prepare(existing, detail map[string]interface{}) map[string]interface{} {
	detail["id"] = existing["id"]
	return detail
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
func (h *DatasourceHandler) GetRepresentation(uid string, detail map[string]interface{}) (string, error) {
	j, err := json.MarshalIndent(detail, "", "  ")
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
func (h *DatasourceHandler) Add(detail map[string]interface{}) error {
	return postDatasource(Datasource(detail))
}

// Update pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Update(existing, detail map[string]interface{}) error {
	detail["id"] = existing["id"]
	return putDatasource(Datasource(detail))
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DatasourceHandler) Preview(detail map[string]interface{}, opts *grizzly.PreviewOpts) error {
	return grizzly.ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////

// getRemoteDatasource retrieves a datasource object from Grafana
func getRemoteDatasource(uid string) (*Datasource, error) {
	grafanaURL, err := getGrafanaURL("api/datasources/name/" + uid)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(grafanaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, grizzly.ErrNotFound
	default:
		if resp.StatusCode >= 400 {
			return nil, errors.New(resp.Status)
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d Datasource
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, APIErr{err, data}
	}
	return &d, nil
}

func postDatasource(source Datasource) error {
	grafanaURL, err := getGrafanaURL("api/datasources")
	if err != nil {
		return err
	}

	sourceJSON, err := source.toJSON()
	if err != nil {
		return err
	}

	resp, err := http.Post(grafanaURL, "application/json", bytes.NewBufferString(sourceJSON))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusPreconditionFailed:
		d := json.NewDecoder(resp.Body)
		var r struct {
			Message string `json:"message"`
		}
		if err := d.Decode(&r); err != nil {
			return fmt.Errorf("Failed to decode actual error (412 Precondition failed): %s", err)
		}
		fmt.Println(sourceJSON)
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", source.UID(), r.Message)
	default:
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s", resp.Status, source.UID())
	}
	return nil
}

func putDatasource(source Datasource) error {
	id, err := source.getID()
	if err != nil {
		return err
	}
	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/datasources/%d", id))
	if err != nil {
		return err
	}

	sourceJSON, err := source.toJSON()
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", grafanaURL, bytes.NewBufferString(sourceJSON))
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusPreconditionFailed:
		d := json.NewDecoder(resp.Body)
		var r struct {
			Message string `json:"message"`
		}
		if err := d.Decode(&r); err != nil {
			return fmt.Errorf("Failed to decode actual error (412 Precondition failed): %s", err)
		}
		fmt.Println(sourceJSON)
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", source.UID(), r.Message)
	default:
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s", resp.Status, source.UID())
	}
	return nil
}

// Datasource encapsulates a datasource
type Datasource map[string]interface{}

// UID retrieves the UID from a datasource
func (d *Datasource) UID() string {
	uid, ok := (*d)["name"]
	if !ok {
		return ""
	}
	return uid.(string)
}

// toJSON returns JSON for a datasource
func (d *Datasource) toJSON() (string, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func (d *Datasource) getID() (int, error) {
	v, ok := (*d)["id"]
	if !ok {
		return 0, fmt.Errorf("Datasource %s requires an ID to update", d.UID())
	}
	id := int(v.(float64))
	return id, nil
}
