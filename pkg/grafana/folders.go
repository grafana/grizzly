package grafana

import (
	"encoding/json"
	"fmt"
	"strings"

	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Losing canAdmin, canDelete, canEdit, canSave, created, createdBy, hasAcl, updated, updatedBy, version

// getRemoteFolder retrieves a folder object from Grafana
func getRemoteFolder(uid string) (*grizzly.Resource, error) {
	h := FolderHandler{}
	var folder *gapi.Folder
	if uid == "General" || uid == "general" {
		folder = &gapi.Folder{
			ID:    0,
			UID:   uid,
			Title: "General",
			// URL: ??
		}
	} else {
		client, err := getClient()
		if err != nil {
			return nil, err
		}

		folder, err = client.FolderByUID(uid)
		if err != nil {
			if strings.HasPrefix(err.Error(), "status: 404") {
				return nil, grizzly.ErrNotFound
			}
			return nil, err
		}
	}

	// TODO: Turn spec into a real gapi.Folder object
	var spec map[string]interface{}
	data, err := json.Marshal(folder)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	return &resource, nil
}

func getRemoteFolderList() ([]string, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	folders, err := client.Folders()
	if err != nil {
		return nil, err
	}

	uids := make([]string, len(folders))
	for i, folder := range folders {
		uids[i] = folder.UID
	}
	return uids, nil
}

func postFolder(resource grizzly.Resource) error {
	name := resource.Name()
	if name == "General" || name == "general" {
		return nil
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	// TODO: Turn spec into a real gapi.Folder object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var folder gapi.Folder
	err = json.Unmarshal(data, &folder)
	if err != nil {
		return err
	}
	if folder.Title == "" {
		return fmt.Errorf("missing title in folder spec")
	}
	_, err = client.NewFolder(folder.Title, folder.UID)
	return err
}

func putFolder(resource grizzly.Resource) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// TODO: Turn spec into a real gapi.Folder object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var folder gapi.Folder
	err = json.Unmarshal(data, &folder)
	if err != nil {
		return err
	}
	if folder.Title == "" {
		return fmt.Errorf("missing title in folder spec")
	}
	return client.UpdateFolder(folder.UID, folder.Title)
}

var getFolderById = func(folderId int64) (*gapi.Folder, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	return client.Folder(folderId)
}
