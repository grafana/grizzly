package grafana

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
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

// Kind returns the name for this handler
func (h *DashboardHandler) Kind() string {
	return "Dashboard"
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *DashboardHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

const (
	dashboardFolderDefault = "General"
)

// GetExtension returns the file name extension for a dashboard
func (h *DashboardHandler) GetExtension() string {
	return "json"
}

func (h *DashboardHandler) newDashboardResource(m manifest.Manifest) grizzly.Resource {
	resource := grizzly.Resource{
		UID:     m.Metadata().Name(),
		Handler: h,
		Detail:  m,
	}
	return resource
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DashboardHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DashboardHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetRemoteByUID retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemoteByUID(uid string) (*grizzly.Resource, error) {
	m, err := getRemoteDashboard(uid)
	if err != nil {
		return nil, err
	}
	return grizzly.NewResource(*m, h), nil
}

// GetRemote retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemote(existing grizzly.Resource) (*grizzly.Resource, error) {
	return h.GetRemoteByUID(existing.Detail.Metadata().Name())
}

// Add pushes a new dashboard to Grafana via the API
func (h *DashboardHandler) Add(resource grizzly.Resource) error {
	return postDashboard(resource.Detail)
}

// Update pushes a dashboard to Grafana via the API
func (h *DashboardHandler) Update(existing, resource grizzly.Resource) error {
	return postDashboard(resource.Detail)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DashboardHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
	s, err := postSnapshot(resource.Detail, opts)
	if err != nil {
		return err
	}
	notifier.Info(&resource, "view: "+s.URL)
	notifier.Error(&resource, "delete: "+s.DeleteURL)
	if opts.ExpiresSeconds > 0 {
		notifier.Warn(&resource, fmt.Sprintf("Previews will expire and be deleted automatically in %d seconds\n", opts.ExpiresSeconds))
	}
	return nil
}

// Listen watches a resource and updates local file on changes
func (h *DashboardHandler) Listen(notifier grizzly.Notifier, UID, filename string) error {
	return watchDashboard(notifier, UID, filename)
}
