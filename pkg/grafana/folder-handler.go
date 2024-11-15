package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

const DefaultFolder = "General"
const DashboardFolderKind = "DashboardFolder"

// FolderHandler is a Grizzly Handler for Grafana dashboard folders
type FolderHandler struct {
	grizzly.BaseHandler
}

// NewFolderHandler returns configuration defining a new Grafana Folder Handler
func NewFolderHandler(provider grizzly.Provider) *FolderHandler {
	return &FolderHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, DashboardFolderKind, false),
	}
}

const (
	folderPattern = "folders/folder-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *FolderHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(folderPattern, resource.Name(), filetype)
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *FolderHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	if !resource.HasSpecString("uid") {
		resource.SetSpecString("uid", resource.Name())
	}
	return &resource
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *FolderHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	for _, key := range []string{"id", "version", "canAdmin", "canDelete", "canEdit", "canSave", "created", "createdBy", "updated", "updatedBy", "url"} {
		resource.DeleteSpecKey(key)
	}
	value := resource.GetSpecValue("parents")
	if value == nil {
		resource.DeleteSpecKey("parents")
	}
	return &resource
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

func (h *FolderHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	uid, ok := resource.GetSpecString("uid")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}
	return uid, nil
}

// Sort sorts according to handler needs
func (h *FolderHandler) Sort(resources grizzly.Resources) grizzly.Resources {
	result := grizzly.NewResources()
	addedToResult := map[string]bool{}
	for _, resource := range resources.AsList() {
		addedToResult[resource.Name()] = false
	}
	for {
		continueLoop := false
		for _, resource := range resources.AsList() {
			if addedToResult[resource.Name()] {
				// already added
				continue
			}
			parentUID, hasParentUID := resource.Spec()["parentUid"]
			// Add root folders
			if !hasParentUID {
				addedToResult[resource.Name()] = true
				result.Add(resource)
				continue
			}
			parentAdded, parentExists := addedToResult[parentUID.(string)]
			// Add folders with parents which aren't declared in Grizzly, or which have already been added
			if !parentExists || parentAdded {
				addedToResult[resource.Name()] = true
				result.Add(resource)
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
func (h *FolderHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteFolder(uid)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard folder %s: %w", uid, err)
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

func (h *FolderHandler) ProxyURL(uid string) string {
	return fmt.Sprintf("/dashboards/f/%s/", uid)
}

func (h *FolderHandler) GetProxyEndpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/alerting/{rule_uid}/edit",
			Handler: authenticateAndProxyHandler(s, h.Provider),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/folders/{folder_uid}",
			Handler: h.FolderJSONGetHandler(s),
		},
	}
}

func (h *FolderHandler) FolderJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folderUID := chi.URLParam(r, "folder_uid")
		withAccessControl := r.URL.Query().Get("accesscontrol")

		folder, found := s.Resources.Find(grizzly.NewResourceRef(DashboardFolderKind, folderUID))
		if !found {
			grizzly.SendError(w, fmt.Sprintf("Folder with UID %s not found", folderUID), fmt.Errorf("folder with UID %s not found", folderUID), http.StatusNotFound)
			return
		}

		// These values are required for the page to load properly.
		if folder.GetSpecValue("version") == nil {
			folder.SetSpecValue("version", 1)
		}
		if folder.GetSpecValue("id") == nil {
			folder.SetSpecValue("id", 1)
		}

		response := folder.Spec()

		if withAccessControl == "true" {
			// TODO: can we omit stuff from this list?
			response["accessControl"] = map[string]any{
				"alert.rules:create":           false,
				"alert.rules:delete":           false,
				"alert.rules:read":             true,
				"alert.rules:write":            false,
				"alert.silences:create":        false,
				"alert.silences:read":          true,
				"alert.silences:write":         false,
				"annotations:create":           false,
				"annotations:delete":           false,
				"annotations:read":             true,
				"annotations:write":            false,
				"dashboards.permissions:read":  true,
				"dashboards.permissions:write": false,
				"dashboards:create":            true,
				"dashboards:delete":            false,
				"dashboards:read":              true,
				"dashboards:write":             true,
				"folders.permissions:read":     true,
				"folders.permissions:write":    false,
				"folders:create":               false,
				"folders:delete":               false,
				"folders:read":                 true,
				"folders:write":                false,
				"library.panels:create":        false,
				"library.panels:delete":        false,
				"library.panels:read":          true,
				"library.panels:write":         false,
			}
		}

		writeJSONOrLog(w, response)
	}
}

// getRemoteFolder retrieves a folder object from Grafana
func (h *FolderHandler) getRemoteFolder(uid string) (*grizzly.Resource, error) {
	if uid == "" {
		return nil, fmt.Errorf("no folder UID provided")
	}
	var folder *models.Folder
	if uid == DefaultFolder || uid == strings.ToLower(DefaultFolder) {
		folder = &models.Folder{
			ID:    0,
			UID:   uid,
			Title: DefaultFolder,
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
				return nil, fmt.Errorf("couldn't fetch folder '%s' from remote: %w", uid, grizzly.ErrNotFound)
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

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (h *FolderHandler) getRemoteFolderList() ([]string, error) {
	var (
		limit            = int64(1000)
		page       int64 = 0
		uids       []string
		folderType = "dash-folder"
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
	if name == DefaultFolder || name == strings.ToLower(DefaultFolder) {
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

	_, err = client.Folders.UpdateFolder(resource.Name(), &body)
	return err
}

var getFolderByID = func(client *gclient.GrafanaHTTPAPI, folderId int64) (*models.Folder, error) {
	folderOk, err := client.Folders.GetFolderByID(folderId)
	if err != nil {
		return nil, err
	}
	return folderOk.GetPayload(), nil
}
