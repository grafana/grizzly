package grafana

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// AlertsHandler is a Grizzly Handler for Prometheus Alerts
type AlertsHandler struct {
	Provider Provider
}

// NewAlertsHandler returns a new Grizzly Handler for Prometheus Alerts
func NewAlertsHandler(provider Provider) *AlertsHandler {
	return &AlertsHandler{
		Provider: provider,
	}
}

// Kind returns the name for this handler
func (h *AlertsHandler) Kind() string {
	return "GrafanaAlertsGroup"
}

// Validate returns the uid of resource
func (h *AlertsHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *AlertsHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a Alerts grouping
func (h *AlertsHandler) GetExtension() string {
	return "yaml"
}

const (
	grafanaAlertsGroupGlob    = "grafana/Alerts-*"
	grafanaAlertsGroupPattern = "grafana/Alerts-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *AlertsHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, grafanaAlertsGroupGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertsHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(grafanaAlertsGroupPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *AlertsHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *AlertsHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *AlertsHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *AlertsHandler) GetUID(resource grizzly.Resource) (string, error) {
	if !resource.HasMetadata("datasource") {
		return "", fmt.Errorf("%s %s requires a namespace metadata entry", h.Kind(), resource.Name())
	}

	return fmt.Sprintf("%s||%s", resource.GetMetadata("datasource"), resource.Name()), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertsHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return getRemoteAlertGroup(uid)
}

// GetRemote retrieves a datasource as a Resource
func (h *AlertsHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid := fmt.Sprintf("%s||%s", resource.GetMetadata("datasource"), resource.Name())
	return getRemoteAlertGroup(uid)
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertsHandler) ListRemote() ([]string, error) {
	return getRemoteAlertGroupList()
}

// Add pushes a datasource to Grafana via the API
func (h *AlertsHandler) Add(resource grizzly.Resource) error {
	return postAlertGroup(resource)
}

// Update pushes a datasource to Grafana via the API
func (h *AlertsHandler) Update(existing, resource grizzly.Resource) error {
	return postAlertGroup(resource)
}
