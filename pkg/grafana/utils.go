package grafana

import "regexp"

var folderURLRegex = regexp.MustCompile("/dashboards/f/([^/]+)")

const generalFolderId = 0

func extractFolderUID(d DashboardWrapper) string {
	folderUid := d.Meta.FolderUID
	if folderUid == "" {
		urlPaths := folderURLRegex.FindStringSubmatch(d.Meta.FolderURL)
		if urlPaths == nil || len(urlPaths) == 0 {
			if d.FolderID == generalFolderId {
				return ""
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
