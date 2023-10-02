package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Moved from utils.go
const generalFolderId = 0
const generalFolderUID = "general"

// Losing id, typeLogoUrl, version, withCredentials

// getRemoteDashboard retrieves a dashboard object from Grafana
func getRemoteDashboard(uid string) (*grizzly.Resource, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	dashboard, err := client.DashboardByUID(uid)
	// TODO: Restore lookup by name functionality, underlying library lacks it
	if err != nil {
		if strings.HasPrefix(err.Error(), "status: 404") {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}

	folderUid := dashboard.FolderUID
	// TODO: Lost attempted parsing of folderURL field
	if folderUid == "" {
		if dashboard.FolderID == generalFolderId {
			folderUid = generalFolderUID
		} else {
			folder, err := getFolderById(dashboard.FolderID)
			if err != nil {
				return nil, err
			}
			folderUid = folder.UID
		}
	}

	// TODO: Turn spec into a real gapi.Dashboard object
	var spec map[string]interface{}
	data, err := json.Marshal(dashboard.Model)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}

	h := DashboardHandler{}
	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	resource.SetMetadata("folder", folderUid)
	return &resource, nil
}

func getRemoteDashboardList() ([]string, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	dashboards, err := client.Dashboards()
	if err != nil {
		return nil, err
	}

	uids := make([]string, len(dashboards))
	for i, dashboard := range dashboards {
		uids[i] = dashboard.UID
	}
	return uids, nil
}

func postDashboard(resource grizzly.Resource) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	folderUID := resource.GetMetadata("folder")
	var folderID int64
	if !(folderUID == "General" || folderUID == "general") {
		folder, err := getRemoteFolder(folderUID)
		if err != nil {
			if errors.Is(err, grizzly.ErrNotFound) {
				return fmt.Errorf("cannot upload dashboard %s as folder %s not found", resource.Name(), folderUID)
			} else {
				return fmt.Errorf("cannot upload dashboard %s: %w", resource.Name(), err)
			}
		}
		folderID = int64(folder.GetSpecValue("id").(float64))
	} else {
		folderID = generalFolderId
	}

	dashboard := gapi.Dashboard{
		Model:     resource.Spec(),
		FolderID:  folderID,
		Overwrite: true,
	}
	_, err = client.NewDashboard(dashboard)
	return err
}

func postSnapshot(resource grizzly.Resource, opts *grizzly.PreviewOpts) (*gapi.SnapshotCreateResponse, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	snapshot := gapi.Snapshot{
		Model: resource.Spec(),
	}
	if opts.ExpiresSeconds > 0 {
		snapshot.Expires = int64(opts.ExpiresSeconds)
	}
	return client.NewSnapshot(snapshot)
}
