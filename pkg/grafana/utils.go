package grafana

import (
	"regexp"

	"github.com/mitchellh/mapstructure"
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

func decode(input, output interface{}) error {
	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &output,
		TagName:  "json",
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	return decoder.Decode(input)
}
