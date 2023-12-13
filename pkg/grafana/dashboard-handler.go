package grafana

import (
	"fmt"
	"path/filepath"

	"errors"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/client/snapshots"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// Moved from utils.go
const generalFolderId = 0
const generalFolderUID = "general"

// DashboardHandler is a Grizzly Handler for Grafana dashboards
type DashboardHandler struct {
	Provider grizzly.Provider
}

// NewDashboardHandler returns configuration defining a new Grafana Dashboard Handler
func NewDashboardHandler(provider grizzly.Provider) *DashboardHandler {
	return &DashboardHandler{
		Provider: provider,
	}
}

// Validate returns the uid of resource
func (h *DashboardHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

// Kind returns the name for this handler
func (h *DashboardHandler) Kind() string {
	return "Dashboard"
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *DashboardHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a dashboard
func (h *DashboardHandler) GetExtension() string {
	return "json"
}

const (
	dashboardGlob    = "dashboards/*/dashboard-*"
	dashboardPattern = "dashboards/%s/dashboard-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *DashboardHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, dashboardGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *DashboardHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(dashboardPattern, resource.GetMetadata("folder"), resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *DashboardHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	resource.SetSpecString("uid", resource.Name())
	if !resource.HasMetadata("folder") {
		resource.SetMetadata("folder", generalFolderUID)
	}
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DashboardHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("id")
	resource.DeleteSpecKey("version")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DashboardHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *DashboardHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DashboardHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteDashboard(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %w", UID, err)
	}
	return resource, nil
}

// GetRemote retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid, _ := resource.GetSpecString("uid")
	if uid != resource.Name() {
		return nil, fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return h.getRemoteDashboard(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *DashboardHandler) ListRemote() ([]string, error) {
	return h.getRemoteDashboardList()
}

// Add pushes a new dashboard to Grafana via the API
func (h *DashboardHandler) Add(resource grizzly.Resource) error {
	resource = *h.Unprepare(resource)
	return h.postDashboard(resource)
}

// Update pushes a dashboard to Grafana via the API
func (h *DashboardHandler) Update(existing, resource grizzly.Resource) error {
	return h.postDashboard(resource)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DashboardHandler) Preview(resource grizzly.Resource, opts *grizzly.PreviewOpts) error {
	s, err := h.postSnapshot(resource, opts)
	if err != nil {
		return err
	}
	notifier.Info(resource, "view: "+s.URL)
	if opts.ExpiresSeconds > 0 {
		notifier.Warn(resource, fmt.Sprintf("Previews will expire and be deleted automatically in %d seconds\n", opts.ExpiresSeconds))
	} else {
		notifier.Error(resource, "delete: "+s.DeleteURL)
	}
	return nil
}

// getRemoteDashboard retrieves a dashboard object from Grafana
func (h *DashboardHandler) getRemoteDashboard(uid string) (*grizzly.Resource, error) {
	params := dashboards.NewGetDashboardByUIDParams().WithUID(uid)
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
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

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	folderUid := extractFolderUID(client, *dashboard)
	resource.SetMetadata("folder", folderUid)
	return &resource, nil
}

func (h *DashboardHandler) getRemoteDashboardList() ([]string, error) {
	var (
		limit            = int64(1000)
		searchType       = "dash-db"
		page       int64 = 0
		uids       []string
	)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

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

func (h *DashboardHandler) postDashboard(resource grizzly.Resource) error {
	folderUID := resource.GetMetadata("folder")
	var folderID int64
	if !(folderUID == "General" || folderUID == "general") {
		folderHandler := NewFolderHandler(h.Provider)
		folder, err := folderHandler.getRemoteFolder(folderUID)
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
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	params := dashboards.NewPostDashboardParams().WithBody(&body)
	_, err = client.Dashboards.PostDashboard(params, nil)
	return err
}

func (h *DashboardHandler) postSnapshot(resource grizzly.Resource, opts *grizzly.PreviewOpts) (*models.CreateDashboardSnapshotOKBody, error) {
	body := models.CreateDashboardSnapshotCommand{
		Dashboard: resource.Spec(),
	}
	if opts.ExpiresSeconds > 0 {
		body.Expires = int64(opts.ExpiresSeconds)
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	params := snapshots.NewCreateDashboardSnapshotParams().WithBody(&body)
	response, err := client.Snapshots.CreateDashboardSnapshot(params, nil)
	if err != nil {
		return nil, err
	}
	return response.GetPayload(), nil
}
