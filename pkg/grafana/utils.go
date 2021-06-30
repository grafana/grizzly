package grafana

import "regexp"

var folderURLRegex = regexp.MustCompile("/dashboards/f/([^/]+)")

func extractFolderUID(d DashboardWrapper) string {
	folderUid := d.Meta.FolderUID
	if folderUid == "" {
		urlPaths := folderURLRegex.FindStringSubmatch(d.Meta.FolderURL)
		if len(urlPaths) == 0 {
			folder, err := getFolderById(int64(d.Dashboard["FolderId"].(float64)))
			if err != nil {
				return ""
			}
			return folder["uid"].(string)
		}
		folderUid = urlPaths[1]
	}
	return folderUid
}
