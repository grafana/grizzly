package grafana

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// FolderHandler is a Grizzly Handler for Grafana dashboard folders
type FolderHandler struct {
	Provider Provider
}

// NewFolderHandler returns configuration defining a new Grafana Folder Handler
func NewFolderHandler(provider Provider) *FolderHandler {
	return &FolderHandler{
		Provider: provider,
	}
}

// Kind returns the name for this handler
func (h *FolderHandler) Kind() string {
	return "DashboardFolder"
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *FolderHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a dashboard
func (h *FolderHandler) GetExtension() string {
	return "json"
}

const (
	folderGlob    = "folders/folder-*"
	folderPattern = "folders/folder-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *FolderHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, folderGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *FolderHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(folderPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *FolderHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	resource.SetSpecString("uid", resource.GetMetadata("name"))
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *FolderHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *FolderHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *FolderHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	resource, err := getRemoteDashboard(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %v", UID, err)
	}
	return resource, nil
}

// GetRemote retrieves a folder as a resource
func (h *FolderHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return getRemoteFolder(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *FolderHandler) ListRemote() ([]string, error) {
	return getRemoteFolderList()
}

// Add pushes a new folder to Grafana via the API
func (h *FolderHandler) Add(resource grizzly.Resource) error {
	return postFolder(resource)
}

// Update pushes a folder to Grafana via the API
func (h *FolderHandler) Update(existing, resource grizzly.Resource) error {
	return putFolder(resource)
}
