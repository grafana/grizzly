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

// DatasourceProvider is a Grizzly Provider for Grafana datasources
type DatasourceProvider struct{}

// NewDatasourceProvider returns configuration defining a new Grafana Provider
func NewDatasourceProvider() *DatasourceProvider {
	return &DatasourceProvider{}
}

// GetName returns the name for this provider
func (p *DatasourceProvider) GetName() string {
	return "grafana"
}

// GetJSONPath returns a paths within Jsonnet output that this provider will consume
func (p *DatasourceProvider) GetJSONPath() string {
	return "grafanaDatasources"
}

// GetExtension returns the file name extension for a datasource
func (p *DatasourceProvider) GetExtension() string {
	return "json"
}

func (p *DatasourceProvider) newDatasourceResource(uid, filename string, board Datasource) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Filename: filename,
		Provider: p,
		Detail:   board,
		Path:     p.GetJSONPath(),
	}
	return resource
}

// Parse parses an interface{} object into a struct for this resource type
func (p *DatasourceProvider) Parse(i interface{}) (grizzly.Resources, error) {
	resources := grizzly.Resources{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		board := Datasource{}
		err := mapstructure.Decode(v, &board)
		if err != nil {
			return nil, err
		}
		resource := p.newDatasourceResource(board.UID(), k, board)
		key := resource.Key()
		resources[key] = resource
	}
	return resources, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (p *DatasourceProvider) GetByUID(UID string) (*grizzly.Resource, error) {
	board, err := getRemoteDatasource(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving datasource %s: %v", UID, err)
	}
	resource := p.newDatasourceResource(UID, "", *board)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (p *DatasourceProvider) GetRepresentation(uid string, detail map[string]interface{}) (string, error) {
	j, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// GetRemoteRepresentation retrieves a datasource as JSON
func (p *DatasourceProvider) GetRemoteRepresentation(uid string) (string, error) {
	board, err := getRemoteDatasource(uid)

	if err != nil {
		return "", err
	}
	return board.toJSON()
}

// GetRemote retrieves a datasource as a Resource
func (p *DatasourceProvider) GetRemote(uid string) (*grizzly.Resource, error) {
	source, err := getRemoteDatasource(uid)
	if err != nil {
		return nil, err
	}
	resource := p.newDatasourceResource(uid, "", *source)
	return &resource, nil
}

// Add pushes a datasource to Grafana via the API
func (p *DatasourceProvider) Add(detail map[string]interface{}) error {
	return postDatasource(Datasource(detail))
}

// Update pushes a datasource to Grafana via the API
func (p *DatasourceProvider) Update(existing, detail map[string]interface{}) error {
	return postDatasource(Datasource(detail))
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (p *DatasourceProvider) Preview(detail map[string]interface{}, opts *grizzly.PreviewOpts) error {
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
	//	delete(d, "version")
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

func putDatasource(url string, source Datasource) error {
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
