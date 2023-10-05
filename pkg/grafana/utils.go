package grafana

import (
	"encoding/json"
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
