package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/grafana/grizzly/pkg/grizzly"
)

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
		return nil, grizzly.APIErr{Err: err, Body: data}
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s %s", source.UID(), resp.Status, string(body))
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s %s", source.UID(), resp.Status, string(body))
	}
	return nil
}

// Datasource encapsulates a datasource
type Datasource map[string]interface{}

func newDatasource(resource grizzly.Resource) Datasource {
	return resource.Detail.(Datasource)
}

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
