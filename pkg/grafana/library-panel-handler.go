package grafana

import (
	"fmt"
	"path/filepath"

	grafana "github.com/grafana/grafana-api-golang-client"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// LibraryPanelHandler is a Grizzly Handler for Grafana library panel
type LibraryPanelHandler struct {
	Provider Provider
}

// NewLibraryPanelHandler returns a new Grizzly Handler for Grafana library panel resources
func NewLibraryPanelHandler(provider Provider) *LibraryPanelHandler {
	return &LibraryPanelHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *LibraryPanelHandler) Kind() string {
	return "LibraryPanel"
}

// Validate returns the uid of resource
func (h *LibraryPanelHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *LibraryPanelHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a library panel
func (h *LibraryPanelHandler) GetExtension() string {
	return "json"
}

const (
	libraryPanelGlob    = "library-panel/library-panel-*"
	libraryPanelPattern = "library-panel/library-panel-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *LibraryPanelHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, libraryPanelGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *LibraryPanelHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(libraryPanelPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *LibraryPanelHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary panels from a remote resource ready for presentation/comparison
func (h *LibraryPanelHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("id")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *LibraryPanelHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *LibraryPanelHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *LibraryPanelHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	panel, err := client.LibraryPanelByUID(UID)
	if err != nil {
		return nil, err
	}

	msi := map[string]interface{}{}
	err = decode(panel, msi)
	if err != nil {
		return nil, err
	}

	handler := LibraryPanelHandler{}
	resource := grizzly.NewResource(handler.APIVersion(), handler.Kind(), UID, msi)
	resource.SetMetadata("folder", panel.Meta.FolderUID)
	return &resource, nil
}

// GetRemote retrieves a datasource as a Resource
func (h *LibraryPanelHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.GetByUID(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *LibraryPanelHandler) ListRemote() ([]string, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	panels, err := client.LibraryPanels()
	if err != nil {
		return nil, err
	}

	UIDs := []string{}
	for _, panel := range panels {
		UIDs = append(UIDs, panel.UID)
	}
	return UIDs, nil
}

// Add pushes a library panel to Grafana via the API
func (h *LibraryPanelHandler) Add(resource grizzly.Resource) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	panel := grafana.LibraryPanel{}
	err = decode(resource.Spec(), &panel)
	if err != nil {
		return err
	}

	_, err = client.NewLibraryPanel(panel)
	return err
}

// Update pushes a library panel to Grafana via the API
func (h *LibraryPanelHandler) Update(existing, resource grizzly.Resource) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	panel := grafana.LibraryPanel{}
	err = decode(resource.Spec(), &panel)
	if err != nil {
		return err
	}

	_, err = client.PatchLibraryPanel(resource.Name(), panel)
	return err
}
