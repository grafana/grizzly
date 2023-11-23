package grafana

import (
	"encoding/json"
	"errors"
	"fmt"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// getRemoteFolder retrieves a folder object from Grafana
func getRemoteFolder(client *gclient.GrafanaHTTPAPI, uid string) (*grizzly.Resource, error) {
	h := FolderHandler{}
	var folder *models.Folder
	if uid == "General" || uid == "general" {
		folder = &models.Folder{
			ID:    0,
			UID:   uid,
			Title: "General",
			// URL: ??
		}
	} else {
		params := folders.NewGetFolderByUIDParams().WithFolderUID(uid)
		folderOk, err := client.Folders.GetFolderByUID(params, nil)
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

func getRemoteFolderList(client *gclient.GrafanaHTTPAPI) ([]string, error) {
	var (
		limit       = int64(1000)
		page  int64 = 0
		uids  []string
	)
	params := folders.NewGetFoldersParams().WithLimit(&limit)
	for {
		page++
		params.SetPage(&page)

		foldersOk, err := client.Folders.GetFolders(params, nil)
		if err != nil {
			return nil, err
		}

		for _, folder := range foldersOk.GetPayload() {
			uids = append(uids, folder.UID)
		}
		if int64(len(foldersOk.GetPayload())) < *params.Limit {
			return uids, nil
		}
	}
}

func postFolder(client *gclient.GrafanaHTTPAPI, resource grizzly.Resource) error {
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
		Title: folder.Title,
		UID:   folder.UID,
	}
	params := folders.NewCreateFolderParams().WithBody(&body)
	_, err = client.Folders.CreateFolder(params, nil)
	return err
}

func putFolder(client *gclient.GrafanaHTTPAPI, resource grizzly.Resource) error {
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
		Title: folder.Title,
	}
	params := folders.NewUpdateFolderParams().WithBody(&body)
	_, err = client.Folders.UpdateFolder(params, nil)
	return err
}

var getFolderById = func(client *gclient.GrafanaHTTPAPI, folderId int64) (*models.Folder, error) {
	params := folders.NewGetFolderByIDParams().WithFolderID(folderId)
	folderOk, err := client.Folders.GetFolderByID(params, nil)
	if err != nil {
		return nil, err
	}
	return folderOk.GetPayload(), nil
}
