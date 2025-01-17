package grafana

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
)

// Moved from utils.go
const generalFolderID = 0
const generalFolderUID = "general"

const DashboardKind = "Dashboard"

var _ grizzly.Handler = &DashboardHandler{}
var _ grizzly.ProxyConfiguratorProvider = &DashboardHandler{}

// DashboardHandler is a Grizzly Handler for Grafana dashboards
type DashboardHandler struct {
	grizzly.BaseHandler
}

// NewDashboardHandler returns configuration defining a new Grafana Dashboard Handler
func NewDashboardHandler(provider grizzly.Provider) *DashboardHandler {
	return &DashboardHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, DashboardKind, true),
	}
}

const (
	dashboardPattern = "dashboards/%s/dashboard-%s.%s"
)

// ProxyConfigurator provides a configurator object describing how to proxy dashboards.
func (h *DashboardHandler) ProxyConfigurator() grizzly.ProxyConfigurator {
	return &dashboardProxyConfigurator{
		provider: h.Provider,
	}
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *DashboardHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(dashboardPattern, resource.GetMetadata("folder"), resource.Name(), filetype)
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DashboardHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("id")
	resource.DeleteSpecKey("version")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DashboardHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	if !resource.HasSpecString("uid") {
		resource.SetSpecString("uid", resource.Name())
	}
	if !resource.HasMetadata("folder") {
		resource.SetMetadata("folder", generalFolderUID)
	}
	return &resource
}

// Validate returns the uid of resource
func (h *DashboardHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if resource.Name() != uid && exist {
		return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return nil
}

func (h *DashboardHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	uid, ok := resource.GetSpecString("uid")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}
	return uid, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DashboardHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteDashboard(uid)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %w", uid, err)
	}
	return resource, nil
}

// GetRemote retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid, _ := resource.GetSpecString("uid")
	if uid != resource.Name() {
		return nil, ErrUIDNameMismatch{UID: uid, Name: resource.Name()}
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
	resource = *h.Unprepare(resource)
	return h.postDashboard(resource)
}

// Snapshot pushes dashboards as snapshots
func (h *DashboardHandler) Snapshot(resource grizzly.Resource, expiresSeconds int) error {
	s, err := h.postSnapshot(resource, expiresSeconds)
	if err != nil {
		return err
	}
	notifier.Info(resource, "view: "+s.URL)
	if expiresSeconds > 0 {
		notifier.Warn(resource, fmt.Sprintf("Snapshots will expire and be deleted automatically in %d seconds\n", expiresSeconds))
	} else {
		notifier.Error(resource, "delete: "+s.DeleteURL)
	}
	return nil
}

// getRemoteDashboard retrieves a dashboard object from Grafana
func (h *DashboardHandler) getRemoteDashboard(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	dashboardOk, err := client.Dashboards.GetDashboardByUID(uid)
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

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	folderUID := extractFolderUID(client, *dashboard)
	resource.SetMetadata("folder", folderUID)
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
	if !(folderUID == DefaultFolder || folderUID == strings.ToLower(DefaultFolder)) {
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
		folderID = generalFolderID
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

	_, err = client.Dashboards.PostDashboard(&body)
	return err
}

func (h *DashboardHandler) postSnapshot(resource grizzly.Resource, expiresSeconds int) (*models.CreateDashboardSnapshotOKBody, error) {
	body := models.CreateDashboardSnapshotCommand{
		Dashboard: resource.Spec(),
	}
	if expiresSeconds > 0 {
		body.Expires = int64(expiresSeconds)
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	response, err := client.Snapshots.CreateDashboardSnapshot(&body, nil)
	if err != nil {
		return nil, err
	}
	return response.GetPayload(), nil
}

func (h *DashboardHandler) Detect(data map[string]any) bool {
	expectedKeys := []string{
		"panels",
		"title",
		"schemaVersion",
	}
	for _, key := range expectedKeys {
		_, ok := data[key]
		if !ok {
			return false
		}
	}
	return true
}
