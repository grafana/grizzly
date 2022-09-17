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

// getRemoteDashboard retrieves a dashboard object from Grafana
func getRemoteDashboard(uid string) (*grizzly.Resource, error) {
	client := new(http.Client)
	grafanaURL, err := getGrafanaURL("api/dashboards/uid/" + uid)
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
		return nil, grizzly.ErrNotFound
	case resp.StatusCode >= 400:
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d DashboardWrapper
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: data}
	}
	delete(d.Dashboard, "id")
	delete(d.Dashboard, "version")
	h := DashboardHandler{}
	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, d.Dashboard)
	folderUid := extractFolderUID(d)
	resource.SetMetadata("folder", folderUid)
	return &resource, nil
}

func getRemoteDashboardList() ([]string, error) {
	batchSize := 500

	client := new(http.Client)
	UIDs := []string{}
	for page := 1; ; page++ {
		grafanaURL, err := getGrafanaURL(fmt.Sprintf("/api/search?type=dash-db&limit=%d&page=%d", batchSize, page))
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
			return nil, grizzly.ErrNotFound
		case resp.StatusCode >= 400:
			return nil, errors.New(resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var dashboards []Dashboard
		if err := json.Unmarshal([]byte(string(body)), &dashboards); err != nil {
			return nil, grizzly.APIErr{Err: err, Body: body}
		}
		for _, dashboard := range dashboards {
			UIDs = append(UIDs, dashboard.UID())
		}
		if len(dashboards) < batchSize {
			break
		}
	}
	return UIDs, nil

}

func postDashboard(resource grizzly.Resource) error {
	client := new(http.Client)
	grafanaURL, err := getGrafanaURL("api/dashboards/db")
	if err != nil {
		return err
	}

	folderUID := resource.GetMetadata("folder")
	var folderID int64
	if !(folderUID == "General" || folderUID == "general") {
		folder, err := getRemoteFolder(folderUID)
		if err != nil {
			if errors.Is(err, grizzly.ErrNotFound) {
				return fmt.Errorf("Cannot upload dashboard %s as folder %s not found", resource.GetMetadata("name"), folderUID)
			} else {
				return fmt.Errorf("Cannot upload dashboard %s: %w", resource.GetMetadata("name"), err)
			}
		}
		folderID = int64(folder.GetSpecValue("id").(float64))
	} else {
		folderID = generalFolderId
	}

	wrappedBoard := DashboardWrapper{
		Dashboard: resource["spec"].(map[string]interface{}),
		FolderID:  folderID,
		Overwrite: true,
	}
	wrappedJSON, err := wrappedBoard.toJSON()

	req, err := http.NewRequest("POST", grafanaURL, bytes.NewBufferString(wrappedJSON))
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
		return nil
	case http.StatusPreconditionFailed:
		d := json.NewDecoder(resp.Body)
		var r struct {
			Message string `json:"message"`
		}
		if err := d.Decode(&r); err != nil {
			return fmt.Errorf("Failed to decode actual error (412 Precondition failed): %s", err)
		}
		fmt.Println(wrappedJSON)
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", resource.Name(), r.Message)
	default:
		return NewErrNon200Response("dashboard", resource.Name(), resp)
	}
}

// SnapshotResp encapsulates the response to a snapshot request
type SnapshotResp struct {
	DeleteKey string `json:"deleteKey"`
	DeleteURL string `json:"deleteUrl"`
	Key       string `json:"key"`
	URL       string `json:"url"`
}

func postSnapshot(resource grizzly.Resource, opts *grizzly.PreviewOpts) (*SnapshotResp, error) {
	client := new(http.Client)
	url, err := getGrafanaURL("api/snapshots")
	if err != nil {
		return nil, err
	}
	type SnapshotReq struct {
		Dashboard map[string]interface{} `json:"dashboard"`
		Expires   int                    `json:"expires,omitempty"`
	}

	sr := &SnapshotReq{
		Dashboard: resource["spec"].(map[string]interface{}),
	}

	if opts.ExpiresSeconds > 0 {
		sr.Expires = opts.ExpiresSeconds
	}

	bs, err := json.Marshal(&sr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, NewErrNon200Response("snapshot", resource.Name(), resp)

	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body: %w", err)
	}

	s := &SnapshotResp{}
	err = json.Unmarshal(b, s)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response body into SnapshotResp: %w", err)
	}
	return s, nil
}

// Dashboard encapsulates a dashboard
type Dashboard map[string]interface{}

// UID retrieves the UID from a dashboard
func (d *Dashboard) UID() string {
	uid, ok := (*d)["uid"]
	if !ok {
		return ""
	}
	return uid.(string)
}

// toJSON returns JSON for a dashboard
func (d *Dashboard) toJSON() (string, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// folderUID retrieves the folder UID for a dashboard
func (d *Dashboard) folderUID() string {
	folderUID, ok := (*d)["folderName"]
	if ok {
		return folderUID.(string)
	}
	return ""
}

// DashboardWrapper adds wrapper to a dashboard JSON. Caters both for Grafana's POST
// API as well as GET which require different JSON.
type DashboardWrapper struct {
	Dashboard Dashboard `json:"dashboard"`
	FolderID  int64     `json:"folderId"`
	Overwrite bool      `json:"overwrite"`
	Meta      struct {
		FolderID    int64  `json:"folderId"`
		FolderTitle string `json:"folderTitle"`
		FolderUID   string `json:"folderUid"`
		FolderURL   string `json:"folderUrl"`
	} `json:"meta"`
}

// UID retrieves the UID from a dashboard wrapper
func (d *DashboardWrapper) UID() string {
	return d.Dashboard.UID()
}

// toJSON returns JSON expected by Grafana API
func (d *DashboardWrapper) toJSON() (string, error) {
	d.Overwrite = true
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}
