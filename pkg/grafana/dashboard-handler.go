package grafana

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// DashboardHandler is a Grizzly Handler for Grafana dashboards
type DashboardHandler struct {
	Provider Provider
}

// NewDashboardHandler returns configuration defining a new Grafana Dashboard Handler
func NewDashboardHandler(provider Provider) *DashboardHandler {
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
	resource, err := getRemoteDashboard(h.Provider.client, UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %v", UID, err)
	}
	return resource, nil
}

// GetRemote retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid, _ := resource.GetSpecString("uid")
	if uid != resource.Name() {
		return nil, fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return getRemoteDashboard(h.Provider.client, resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *DashboardHandler) ListRemote() ([]string, error) {
	return getRemoteDashboardList(h.Provider.client)
}

// Add pushes a new dashboard to Grafana via the API
func (h *DashboardHandler) Add(resource grizzly.Resource) error {
	return postDashboard(h.Provider.client, resource)
}

// Update pushes a dashboard to Grafana via the API
func (h *DashboardHandler) Update(existing, resource grizzly.Resource) error {
	return postDashboard(h.Provider.client, resource)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DashboardHandler) Preview(resource grizzly.Resource, opts *grizzly.PreviewOpts) error {
	s, err := postSnapshot(h.Provider.client, resource, opts)
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
