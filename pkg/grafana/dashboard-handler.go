package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/kylelemons/godebug/diff"
	"github.com/mitchellh/mapstructure"
)

/*
 * This DashboardHandler supports folders. Add a `folderName` to your dashboard JSON.
 * This will be removed from the JSON, and if no folder exists, a dashboard folder
 * will be created with UID and title matching your `folderName`.
 *
 */

// DashboardHandler is a Grizzly Provider for Grafana dashboards
type DashboardHandler struct{}

// NewDashboardHandler returns configuration defining a new Grafana Provider
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// GetName returns the name for this provider
func (h *DashboardHandler) GetName() string {
	return "dashboard"
}

// GetFullName returns the name for this provider
func (h *DashboardHandler) GetFullName() string {
	return "grafana.dashboard"
}

// GetJSONPath returns a paths within Jsonnet output that this provider will consume
func (h *DashboardHandler) GetJSONPath() string {
	return "grafanaDashboards"
}

// GetExtension returns the file name extension for a dashboard
func (h *DashboardHandler) GetExtension() string {
	return "json"
}

func (h *DashboardHandler) newDashboardResource(uid, filename string, board Dashboard) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Filename: filename,
		Handler:  h,
		Detail:   board,
		Path:     h.GetJSONPath(),
	}
	return resource
}

// Parse parses an interface{} object into a struct for this resource type
func (h *DashboardHandler) Parse(i interface{}) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		board := Dashboard{}
		err := mapstructure.Decode(v, &board)
		if err != nil {
			return nil, err
		}
		resource := h.newDashboardResource(board.UID(), k, board)
		key := resource.Key()
		resources[key] = resource
	}
	return resources, nil
}

// Diff compares local resources with remote equivalents and output result
func (h *DashboardHandler) Diff(notifier grizzly.Notifier, resources grizzly.ResourceList) error {
	for _, resource := range resources {
		local, err := resource.GetRepresentation()
		if err != nil {
			return nil
		}
		resource = *h.Unprepare(resource)
		uid := resource.UID
		remote, err := h.GetRemote(resource.UID)
		if err == grizzly.ErrNotFound {
			notifier.NotFound(resource)
			continue
		}
		if err != nil {
			return fmt.Errorf("Error retrieving resource from %s %s: %v", resource.Kind(), uid, err)
		}
		remote = h.Unprepare(*remote)
		remoteRepresentation, err := (*remote).GetRepresentation()
		if err != nil {
			return err
		}

		if local == remoteRepresentation {
			notifier.NoChanges(resource)
		} else {
			difference := diff.Diff(remoteRepresentation, local)
			notifier.HasChanges(resource, difference)
		}
	}
	return nil
}

// Apply local resources to remote endpoint
func (h *DashboardHandler) Apply(notifier grizzly.Notifier, resources grizzly.ResourceList) error {
	for _, resource := range resources {
		existingResource, err := h.GetRemote(resource.UID)
		if err == grizzly.ErrNotFound {

			err := h.Add(resource)
			if err != nil {
				return err
			}
			notifier.Added(resource)
			continue
		} else if err != nil {
			return err
		}
		resourceRepresentation, err := resource.GetRepresentation()
		if err != nil {
			return err
		}
		resource = *h.Prepare(*existingResource, resource)
		existingResource = h.Unprepare(*existingResource)
		existingResourceRepresentation, err := existingResource.GetRepresentation()
		if err != nil {
			return nil
		}
		if resourceRepresentation == existingResourceRepresentation {
			notifier.NoChanges(resource)
		} else {
			err = h.Update(*existingResource, resource)
			if err != nil {
				return err
			}
			notifier.Updated(resource)
		}
	}
	return nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DashboardHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DashboardHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DashboardHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	board, err := getRemoteDashboard(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %v", UID, err)
	}
	resource := h.newDashboardResource(UID, "", *board)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (h *DashboardHandler) GetRepresentation(uid string, resource grizzly.Resource) (string, error) {
	j, err := json.MarshalIndent(resource.Detail, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// GetRemoteRepresentation retrieves a dashboard as JSON
func (h *DashboardHandler) GetRemoteRepresentation(uid string) (string, error) {
	board, err := getRemoteDashboard(uid)

	if err != nil {
		return "", err
	}
	return board.toJSON()
}

// GetRemote retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemote(uid string) (*grizzly.Resource, error) {
	board, err := getRemoteDashboard(uid)
	if err != nil {
		return nil, err
	}
	resource := h.newDashboardResource(uid, "", *board)
	return &resource, nil
}

// Add pushes a new dashboard to Grafana via the API
func (h *DashboardHandler) Add(resource grizzly.Resource) error {
	board := newDashboard(resource)

	if err := postDashboard(board); err != nil {
		return err
	}
	return nil
}

// Update pushes a dashboard to Grafana via the API
func (h *DashboardHandler) Update(existing, resource grizzly.Resource) error {
	board := newDashboard(resource)

	return postDashboard(board)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DashboardHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
	board := newDashboard(resource)
	s, err := postSnapshot(board, opts)
	if err != nil {
		return err
	}
	notifier.Info(resource, "view: "+s.URL)
	notifier.Error(resource, "delete: "+s.DeleteURL)
	if opts.ExpiresSeconds > 0 {
		notifier.Warn(resource, fmt.Sprintf("Previews will expire and be deleted automatically in %d seconds\n", opts.ExpiresSeconds))
	}
	return nil
}
