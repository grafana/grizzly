package grafana

import (
	"encoding/json"
	"regexp"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/models"
)

var folderURLRegex = regexp.MustCompile("/dashboards/f/([^/]+)")

func extractFolderUID(client *gclient.GrafanaHTTPAPI, d models.DashboardFullWithMeta) string {
	folderUid := d.Meta.FolderUID
	if folderUid == "" {
		urlPaths := folderURLRegex.FindStringSubmatch(d.Meta.FolderURL)
		if len(urlPaths) == 0 {
			if d.Meta.FolderID == generalFolderId { // nolint:staticcheck
				return generalFolderUID
			}
			folder, err := getFolderById(client, d.Meta.FolderID) // nolint:staticcheck
			if err != nil {
				return ""
			}
			return folder.UID
		}
		folderUid = urlPaths[1]
	}
	return folderUid
}

func structToMap(s interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return result, nil
}
