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

func makeDatasourceRequest(url string) ([]byte, error) {
	grafanaURL, err := getGrafanaURL(url)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(grafanaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, grizzly.ErrNotFound
	case resp.StatusCode >= 400:
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// getRemoteDatasource retrieves a datasource object from Grafana
func getRemoteDatasource(uid string) (*grizzly.Resource, error) {
	data, err := makeDatasourceRequest("api/datasources/uid/" + uid)
	if errors.Is(err, grizzly.ErrNotFound) {
		data, err = makeDatasourceRequest("api/datasources/name/" + uid)
	}
	if err != nil {
		return nil, err
	}

	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: data}
	}
	handler := DatasourceHandler{}
	resource := grizzly.NewResource(handler.APIVersion(), handler.Kind(), uid, d)
	return &resource, nil
}

func getRemoteDatasourceList() ([]string, error) {
	grafanaURL, err := getGrafanaURL("api/datasources")
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(grafanaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, grizzly.ErrNotFound
	case resp.StatusCode >= 400:
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var datasources []map[string]interface{}
	if err := json.Unmarshal([]byte(string(body)), &datasources); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: body}
	}
	UIDs := []string{}
	for _, datasource := range datasources {
		UID := datasource["uid"].(string)
		UIDs = append(UIDs, UID)
	}
	return UIDs, nil
}

func postDatasource(resource grizzly.Resource) error {
	grafanaURL, err := getGrafanaURL("api/datasources")
	if err != nil {
		return err
	}

	sourceJSON, err := resource.SpecAsJSON()
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
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", resource.Name(), r.Message)
	default:
		return NewErrNon200Response("datasource", resource.Name(), resp)
	}
	return nil
}

func putDatasource(resource grizzly.Resource) error {
	spec := resource.Spec()
	id := int64(spec["id"].(float64))
	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/datasources/%d", id))
	if err != nil {
		return err
	}

	sourceJSON, err := resource.SpecAsJSON()
	if err != nil {
		return err
	}

	client, err := NewHttpClient()
	if err != nil {
		return err
	}
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
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", resource.Name(), r.Message)
	default:
		return NewErrNon200Response("datasource", resource.Name(), resp)
	}
	return nil
}
