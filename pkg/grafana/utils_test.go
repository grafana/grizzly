package grafana

import (
	"testing"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestExtractFolderUID(t *testing.T) {
	provider := NewProvider()

	client, err := provider.Client()
	require.NoError(t, err)

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{client: client},
		})

	t.Run("extract folder uid successfully - uid exists", func(t *testing.T) {
		meta := models.DashboardMeta{
			FolderUID: "sample",
		}
		dashboard := models.DashboardFullWithMeta{
			Meta: &meta,
		}
		uid := extractFolderUID(client, dashboard)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid successfully - url exists", func(t *testing.T) {
		meta := models.DashboardMeta{
			FolderURL: "/dashboards/f/sample/special-sample-folder",
		}
		dashboard := models.DashboardFullWithMeta{
			Meta: &meta,
		}
		uid := extractFolderUID(client, dashboard)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid - empty uid returned", func(t *testing.T) {
		meta := models.DashboardMeta{
			FolderID: 1,
		}
		dashboard := models.DashboardFullWithMeta{
			Meta: &meta,
		}
		getFolderById = func(client *gclient.GrafanaHTTPAPI, folderId int64) (*models.Folder, error) {
			return &models.Folder{
				UID: "12345",
			}, nil
		}
		uid := extractFolderUID(client, dashboard)
		require.Equal(t, "12345", uid)
	})
}
