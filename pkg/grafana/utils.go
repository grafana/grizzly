package grafana

import (
	"regexp"
)

var folderURLRegex = regexp.MustCompile("/dashboards/f/([^/]+)")

const generalFolderId = 0
const generalFolderUID = "general"

func extractFolderUID(d DashboardWrapper) string {
	folderUid := d.Meta.FolderUID
	if folderUid == "" {
		urlPaths := folderURLRegex.FindStringSubmatch(d.Meta.FolderURL)
		if len(urlPaths) == 0 {
			if d.FolderID == generalFolderId {
				return generalFolderUID
			}
			folder, err := getFolderById(d.FolderID)
			if err != nil {
				return ""
			}
			return folder["uid"].(string)
		}
		folderUid = urlPaths[1]
	}
	return folderUid
}
