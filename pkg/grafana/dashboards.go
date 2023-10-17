package grafana

import (
	"errors"
	"fmt"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/client/snapshots"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/grafana/grizzly/pkg/grizzly"
)

// Moved from utils.go
const generalFolderId = 0
const generalFolderUID = "general"

// Losing id, typeLogoUrl, version, withCredentials

// getRemoteDashboard retrieves a dashboard object from Grafana
func getRemoteDashboard(client *gclient.GrafanaHTTPAPI, uid string) (*grizzly.Resource, error) {
	params := dashboards.NewGetDashboardByUIDParams().WithUID(uid)
	dashboardOk, err := client.Dashboards.GetDashboardByUID(params, nil)
	if err != nil {
		var gErr *dashboards.GetDashboardByUIDNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}
	dashboard := dashboardOk.GetPayload()

	// TODO: Turn spec into a real models.DashboardFullWithMeta object
	spec, err := structToMap(dashboard.Dashboard)
	if err != nil {
		return nil, err
	}

	h := DashboardHandler{}
	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	folderUid := extractFolderUID(client, *dashboard)
	resource.SetMetadata("folder", folderUid)
	return &resource, nil
}

func getRemoteDashboardList(client *gclient.GrafanaHTTPAPI) ([]string, error) {
	var (
		limit            = int64(1000)
		searchType       = "dash-db"
		page       int64 = 0
		uids       []string
	)

	params := search.NewSearchParams().WithLimit(&limit).WithType(&searchType)
	for {
		page++
		params.SetPage(&page)

		searchOk, err := client.Search.Search(params, nil)
		if err != nil {
			return nil, err
		}

		for _, hit := range searchOk.GetPayload() {
			uids = append(uids, hit.UID)
		}
		if int64(len(searchOk.GetPayload())) < *params.Limit {
			return uids, nil
		}
	}
}

func postDashboard(client *gclient.GrafanaHTTPAPI, resource grizzly.Resource) error {
	folderUID := resource.GetMetadata("folder")
	var folderID int64
	if !(folderUID == "General" || folderUID == "general") {
		folder, err := getRemoteFolder(client, folderUID)
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

	body := models.SaveDashboardCommand{
		Dashboard: resource.Spec(),
		FolderID:  folderID,
		Overwrite: true,
	}
	params := dashboards.NewPostDashboardParams().WithBody(&body)
	_, err := client.Dashboards.PostDashboard(params, nil)
	return err
}

func postSnapshot(client *gclient.GrafanaHTTPAPI, resource grizzly.Resource, opts *grizzly.PreviewOpts) (*models.CreateDashboardSnapshotOKBody, error) {
	body := models.CreateDashboardSnapshotCommand{
		Dashboard: resource.Spec(),
	}
	if opts.ExpiresSeconds > 0 {
		body.Expires = int64(opts.ExpiresSeconds)
	}
	params := snapshots.NewCreateDashboardSnapshotParams().WithBody(&body)
	response, err := client.Snapshots.CreateDashboardSnapshot(params, nil)
	if err != nil {
		return nil, err
	}
	return response.GetPayload(), nil
}
