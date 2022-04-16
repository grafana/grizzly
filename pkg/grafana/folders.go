package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/grafana/grizzly/pkg/grizzly"
)

// getRemoteFolder retrieves a folder object from Grafana
func getRemoteFolder(uid string) (*grizzly.Resource, error) {
	client := new(http.Client)
	h := FolderHandler{}
	if uid == "General" || uid == "general" {
		folder := Folder{
			"id":    0.0,
			"uid":   uid,
			"title": "General",
		}
		resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, folder)
		return &resource, nil
	}
	grafanaURL, err := getGrafanaURL("api/folders/" + uid)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", grafanaURL, nil)
	if err != nil {
		return nil, err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("couldn't fetch folder '%s' from remote: %w", uid, grizzly.ErrNotFound)
	case resp.StatusCode >= 400:
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var f Folder
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: data}
	}
	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, f)
	return &resource, nil
}

func getRemoteFolderList() ([]string, error) {
	batchSize := 100

	client := new(http.Client)

	UIDs := []string{}
	for page := 1; ; page++ {
		grafanaURL, err := getGrafanaURL(fmt.Sprintf("/api/search?type=dash-folder&limit=%d&page=%d", batchSize, page))
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("GET", grafanaURL, nil)
		if err != nil {
			return nil, err
		}

		if grafanaToken, ok := getGrafanaToken(); ok {
			req.Header.Set("Authorization", "Bearer "+grafanaToken)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch {
		case resp.StatusCode == http.StatusNotFound:
			return nil, fmt.Errorf("couldn't fetch folder list from remote: %w", grizzly.ErrNotFound)
		case resp.StatusCode >= 400:
			return nil, errors.New(resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var folders []Folder
		if err := json.Unmarshal([]byte(string(body)), &folders); err != nil {
			return nil, grizzly.APIErr{Err: err, Body: body}
		}
		for _, folder := range folders {
			UIDs = append(UIDs, folder.UID())
		}
		if len(folders) < batchSize {
			break
		}
	}
	return UIDs, nil

}

func postFolder(resource grizzly.Resource) error {
	name := resource.GetMetadata("name")
	client := new(http.Client)
	if name == "General" || name == "general" {
		return nil
	}
	grafanaURL, err := getGrafanaURL("api/folders")
	if err != nil {
		return err
	}

	folder := Folder(resource["spec"].(map[string]interface{}))
	folder["uid"] = resource.GetMetadata("name")
	folderJSON, err := folder.toJSON()

	req, err := http.NewRequest("POST", grafanaURL, bytes.NewBufferString(folderJSON))
	if err != nil {
		return err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}
	req.Header.Set("Content-Type", "application/json")

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
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", resource.Name(), r.Message)
	default:
		return NewErrNon200Response("folder", resource.Name(), resp)
	}
	return nil
}

func putFolder(resource grizzly.Resource) error {
	uid := resource.GetMetadata("name")
	grafanaURL, err := getGrafanaURL("api/folders/" + uid)
	if err != nil {
		return err
	}

	folder := Folder(resource["spec"].(map[string]interface{}))
	folder["overwrite"] = true
	folderJSON, err := folder.toJSON()
	req, err := http.NewRequest(http.MethodPut, grafanaURL, bytes.NewBufferString(folderJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
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
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", resource.Name(), r.Message)
	default:
		return NewErrNon200Response("folder", resource.Name(), resp)
	}

	return nil
}

var getFolderById = func(folderId int64) (Folder, error) {
	client := new(http.Client)
	grafanaURL, err := getGrafanaURL(fmt.Sprintf("folders/id/%d", folderId))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", grafanaURL, nil)
	if err != nil {
		return nil, err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("NOT FOUND")
	default:
		if resp.StatusCode >= 400 {
			return nil, errors.New(resp.Status)
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var f Folder
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f, nil
}

type Folder map[string]interface{}

func (f *Folder) UID() string {
	return (*f)["uid"].(string)
}

func (f *Folder) ID() float64 {
	return (*f)["id"].(float64)
}

// toJSON returns JSON expected by Grafana API
func (f *Folder) toJSON() (string, error) {
	j, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}
