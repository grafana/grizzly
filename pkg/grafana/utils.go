package grafana

import (
	"regexp"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
)

var folderURLRegex = regexp.MustCompile("/dashboards/f/([^/]+)")

const generalFolderId = 0
const generalFolderUID = "general"

func extractFolderUID(client *gclient.GrafanaHTTPAPI, d DashboardWrapper) string {
	folderUid := d.Meta.FolderUID
	if folderUid == "" {
		urlPaths := folderURLRegex.FindStringSubmatch(d.Meta.FolderURL)
		if len(urlPaths) == 0 {
			if d.FolderID == generalFolderId {
				return generalFolderUID
			}
			folder, err := getFolderById(client, d.FolderID)
			if err != nil {
				return ""
			}
			return folder["uid"].(string)
		}
		folderUid = urlPaths[1]
	}
	return folderUid
}
