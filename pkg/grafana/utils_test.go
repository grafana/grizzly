package grafana

import (
	"os"
	"testing"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grizzly/pkg/grizzly"
	. "github.com/grafana/grizzly/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestExtractFolderUID(t *testing.T) {
	os.Setenv("GRAFANA_URL", GetUrl())

	client, err := GetClient()
	require.NoError(t, err)

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{client: client},
		})

	t.Run("extract folder uid successfully - uid exists", func(t *testing.T) {
		dashboardWrapper := DashboardWrapper{}
		dashboardWrapper.Meta.FolderUID = "sample"
		uid := extractFolderUID(client, dashboardWrapper)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid successfully - url exists", func(t *testing.T) {
		dashboardWrapper := DashboardWrapper{}
		url := "/dashboards/f/sample/special-sample-folder"
		dashboardWrapper.Meta.FolderURL = url
		uid := extractFolderUID(client, dashboardWrapper)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid - empty uid returned", func(t *testing.T) {
		dashboardWrapper := DashboardWrapper{
			FolderID: 1,
		}
		getFolderById = func(client *gclient.GrafanaHTTPAPI, folderId int64) (Folder, error) {
			return Folder{
				"uid": "12345",
			}, nil
		}
		uid := extractFolderUID(client, dashboardWrapper)
		require.Equal(t, "12345", uid)
	})
}
