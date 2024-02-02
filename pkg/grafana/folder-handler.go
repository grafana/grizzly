package grafana

import (
	"fmt"
	"path/filepath"

	"encoding/json"
	"errors"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// FolderHandler is a Grizzly Handler for Grafana dashboard folders
type FolderHandler struct {
	Provider grizzly.Provider
}

// NewFolderHandler returns configuration defining a new Grafana Folder Handler
func NewFolderHandler(provider grizzly.Provider) *FolderHandler {
	return &FolderHandler{
		Provider: provider,
	}
}

// Kind returns the name for this handler
func (h *FolderHandler) Kind() string {
	return "DashboardFolder"
}

// Validate returns the uid of resource
func (h *FolderHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}

	return nil
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *FolderHandler) APIVersion() string {
	return h.Provider.APIVersion()
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
	resource, err := grizzly.ResourceFromMap(m)
	if err != nil {
		return nil, err
	}
	resource.SetSpecString("uid", resource.Name())
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

// GetUID returns the UID for a resource
func (h *FolderHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// Sort sorts according to handler needs
func (h *FolderHandler) Sort(resources grizzly.Resources) grizzly.Resources {
	result := grizzly.Resources{}
	addedToResult := map[string]bool{}
	for _, resource := range resources {
		addedToResult[resource.UID()] = false
	}
	for {
		continueLoop := false
		for _, resource := range resources {
			if addedToResult[resource.UID()] {
				// already added
				continue
			}
			parentUID, hasParentUID := resource.Spec()["parentUid"]
			// Add root folders
			if !hasParentUID {
				addedToResult[resource.UID()] = true
				result = append(result, resource)
				continue
			}
			parentAdded, parentExists := addedToResult[parentUID.(string)]
			// Add folders with parents which aren't declared in Grizzly, or which have already been added
			if !parentExists || parentAdded {
				addedToResult[resource.UID()] = true
				result = append(result, resource)
				continue
			}

			// Delay folders with parents which haven't been added yet
			continueLoop = true
		}

		if !continueLoop {
			break
		}
	}

	return result
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *FolderHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteFolder(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard folder %s: %w", UID, err)
	}

	return resource, nil
}

// GetRemote retrieves a folder as a resource
func (h *FolderHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteFolder(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *FolderHandler) ListRemote() ([]string, error) {
	return h.getRemoteFolderList()
}

// Add pushes a new folder to Grafana via the API
func (h *FolderHandler) Add(resource grizzly.Resource) error {
	return h.postFolder(resource)
}

// Update pushes a folder to Grafana via the API
func (h *FolderHandler) Update(existing, resource grizzly.Resource) error {
	return h.putFolder(resource)
}

// getRemoteFolder retrieves a folder object from Grafana
func (h *FolderHandler) getRemoteFolder(uid string) (*grizzly.Resource, error) {
	var folder *models.Folder
	if uid == "General" || uid == "general" {
		folder = &models.Folder{
			ID:    0,
			UID:   uid,
			Title: "General",
			// URL: ??
		}
	} else {
		client, err := h.Provider.(ClientProvider).Client()
		if err != nil {
			return nil, err
		}

		folderOk, err := client.Folders.GetFolderByUID(uid)
		if err != nil {
			var gErrNotFound *folders.GetFolderByUIDNotFound
			var gErrForbidden *folders.GetFolderByUIDForbidden
			if errors.As(err, &gErrNotFound) || errors.As(err, &gErrForbidden) {
				return nil, fmt.Errorf("Couldn't fetch folder '%s' from remote: %w", uid, grizzly.ErrNotFound)
			}
			return nil, err
		}
		folder = folderOk.GetPayload()
	}

	// TODO: Turn spec into a real models.Folder object
	spec, err := structToMap(folder)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	return &resource, nil
}

func (h *FolderHandler) getRemoteFolderList() ([]string, error) {
	var (
		limit            = int64(1000)
		page       int64 = 0
		uids       []string
		folderType string = "dash-folder"
	)

	params := search.NewSearchParams().WithLimit(&limit)
	params.Type = &folderType
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	for {
		page++
		params.SetPage(&page)

		searchOk, err := client.Search.Search(params, nil)
		if err != nil {
			return nil, err
		}

		for _, folder := range searchOk.GetPayload() {
			uids = append(uids, folder.UID)
		}
		if int64(len(searchOk.GetPayload())) < *params.Limit {
			return uids, nil
		}
	}
}

func (h *FolderHandler) postFolder(resource grizzly.Resource) error {
	name := resource.Name()
	if name == "General" || name == "general" {
		return nil
	}

	// TODO: Turn spec into a real models.Folder object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var folder models.Folder
	err = json.Unmarshal(data, &folder)
	if err != nil {
		return err
	}
	if folder.Title == "" {
		return fmt.Errorf("missing title in folder spec")
	}

	body := models.CreateFolderCommand{
		Title:     folder.Title,
		UID:       folder.UID,
		ParentUID: folder.ParentUID,
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	_, err = client.Folders.CreateFolder(&body, nil)
	return err
}

func (h *FolderHandler) putFolder(resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.Folder object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var folder models.Folder
	err = json.Unmarshal(data, &folder)
	if err != nil {
		return err
	}
	if folder.Title == "" {
		return fmt.Errorf("missing title in folder spec")
	}

	body := models.UpdateFolderCommand{
		Title:     folder.Title,
		Overwrite: true,
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	_, err = client.Folders.UpdateFolder(resource.UID(), &body)
	return err
}

var getFolderById = func(client *gclient.GrafanaHTTPAPI, folderId int64) (*models.Folder, error) {
	folderOk, err := client.Folders.GetFolderByID(folderId)
	if err != nil {
		return nil, err
	}
	return folderOk.GetPayload(), nil
}
