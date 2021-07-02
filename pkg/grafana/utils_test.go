package grafana

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExtractFolderUID(t *testing.T) {
	t.Run("extract folder uid successfully - uid exists", func(t *testing.T) {
		dashboardWrapper := DashboardWrapper{}
		dashboardWrapper.Meta.FolderUID = "sample"
		uid := extractFolderUID(dashboardWrapper)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid successfully - url exists", func(t *testing.T) {
		dashboardWrapper := DashboardWrapper{}
		url := "/dashboards/f/sample/special-sample-folder"
		dashboardWrapper.Meta.FolderURL = url
		uid := extractFolderUID(dashboardWrapper)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid - empty uid returned", func(t *testing.T) {
		dashboardWrapper := DashboardWrapper{
			FolderID: 1,
		}
		getFolderById = func(folderId int64) (Folder, error) {
			return Folder{
				"uid": "12345",
			}, nil
		}
		uid := extractFolderUID(dashboardWrapper)
		require.Equal(t, "12345", uid)
	})
}
