package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/manifests"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

const folderNameField = "folderName"

// getRemoteDashboard retrieves a dashboard object from Grafana
func getRemoteDashboard(uid string) (*manifest.Manifest, error) {
	grafanaURL, err := getGrafanaURL("api/dashboards/uid/" + uid)
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

	var d DashboardWrapper
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: data}
	}
	m, err := manifests.New("Dashboard", uid, nil, d.Dashboard)
	m = manifests.RemoveSpecFields(m, []string{"id", "version", "uid"})
	if err != nil {
		return nil, err
	}
	metadata := map[string]interface{}(m.Metadata())
	metadata["folder"] = d.Meta.FolderTitle
	(*m)["metadata"] = metadata
	return m, nil
}

func postDashboard(m manifest.Manifest) error {
	name := m.Metadata().Name()
	grafanaURL, err := getGrafanaURL("api/dashboards/db")
	if err != nil {
		return err
	}

	folderUID := m.Metadata()["folder"].(string)
	folderID, err := findOrCreateFolder(folderUID)
	if err != nil {
		return err
	}
	dashboard := m["spec"].(map[string]interface{})
	dashboard["uid"] = name
	wrappedBoard := DashboardWrapper{
		Dashboard: dashboard,
		FolderID:  folderID,
		Overwrite: true,
	}
	wrappedJSON, err := wrappedBoard.toJSON()

	resp, err := http.Post(grafanaURL, "application/json", bytes.NewBufferString(wrappedJSON))
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
		fmt.Println(wrappedJSON)
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", name, r.Message)
	default:
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s", resp.Status, name)
	}

	return nil
}

// SnapshotResp encapsulates the response to a snapshot request
type SnapshotResp struct {
	DeleteKey string `json:"deleteKey"`
	DeleteURL string `json:"deleteUrl"`
	Key       string `json:"key"`
	URL       string `json:"url"`
}

func postSnapshot(m manifest.Manifest, opts *grizzly.PreviewOpts) (*SnapshotResp, error) {

	url, err := getGrafanaURL("api/snapshots")
	if err != nil {
		return nil, err
	}
	type SnapshotReq struct {
		Dashboard map[string]interface{} `json:"dashboard"`
		Expires   int                    `json:"expires,omitempty"`
	}

	sr := &SnapshotReq{
		Dashboard: m["data"].(map[string]interface{}),
	}

	if opts.ExpiresSeconds > 0 {
		sr.Expires = opts.ExpiresSeconds
	}

	bs, err := json.Marshal(&sr)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response from Grafana: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
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

// DashboardWrapper adds wrapper to a dashboard JSON. Caters both for Grafana's POST
// API as well as GET which require different JSON.
type DashboardWrapper struct {
	Dashboard map[string]interface{} `json:"dashboard"`
	FolderID  int64                  `json:"folderId"`
	Overwrite bool                   `json:"overwrite"`
	Meta      struct {
		FolderID    int64  `json:"folderId"`
		FolderTitle string `json:"folderTitle"`
	} `json:"meta"`
}

// UID retrieves the UID from a dashboard wrapper
func (d *DashboardWrapper) UID() string {
	return d.Dashboard["uid"].(string)
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

// Folder encapsulates a dashboard folder object from the Grafana API
type Folder struct {
	ID    int64  `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

// toJSON returns JSON expected by Grafana API
func (f *Folder) toJSON() (string, error) {
	j, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func findOrCreateFolder(UID string) (int64, error) {
	if UID == "0" || UID == "" || UID == dashboardFolderDefault {
		return 0, nil
	}
	grafanaURL, err := getGrafanaURL("api/folders/" + UID)
	if err != nil {
		return 0, err
	}
	resp, err := http.Get(grafanaURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		var folder Folder
		if err := json.Unmarshal([]byte(string(body)), &folder); err != nil {
			return 0, err
		}
		return folder.ID, nil

	} else if resp.StatusCode == 404 {
		return createFolder(UID)

	} else {
		return 0, fmt.Errorf("Getting folder %s returned error %d", UID, resp.StatusCode)
	}
}

func createFolder(UID string) (int64, error) {
	grafanaURL, err := getGrafanaURL("api/folders")
	if err != nil {
		return 0, err
	}
	folder := Folder{
		UID:   UID,
		Title: UID,
	}

	folderJSON, err := folder.toJSON()
	if err != nil {
		return 0, err
	}
	resp, err := http.Post(grafanaURL, "application/json", bytes.NewBufferString(folderJSON))
	if err != nil {
		return 0, err
	} else if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("Non-200 response from Grafana while applying folder %s: %s", UID, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal([]byte(string(body)), &folder); err != nil {
		return 0, err
	}

	return folder.ID, nil
}
