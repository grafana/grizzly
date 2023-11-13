package grafana

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
	. "github.com/grafana/grizzly/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestDashboard(t *testing.T) {
	os.Setenv("GRAFANA_URL", GetUrl())

	grafanaClient, err := GetClient()
	require.NoError(t, err)

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			NewProvider(grafanaClient),
		})

	ticker := PingService(GetUrl())
	defer ticker.Stop()

	handler, err := grizzly.Registry.GetHandler((&DashboardHandler{}).Kind())
	require.NoError(t, err)

	t.Run("get remote dashboard - success", func(t *testing.T) {
		resource, err := handler.GetByUID("ReciqtgGk")
		require.NoError(t, err)

		require.Equal(t, "grizzly.grafana.com/v1alpha1", resource.APIVersion())
		require.Equal(t, "ReciqtgGk", resource.Name())
		require.NotEmpty(t, resource.GetMetadata("folder"))
	})

	t.Run("get remote dashboard - not found", func(t *testing.T) {
		_, err := handler.GetByUID("dummy")
		require.ErrorContains(t, err, "not found")
	})

	t.Run("get remote dashboard list - success", func(t *testing.T) {
		list, err := handler.ListRemote()
		require.NoError(t, err)

		require.Len(t, list, 3)
		require.EqualValues(t, []string{"ReciqtgGk", "392Ik4GGk", "kE0IIVGGz"}, list)
	})

	t.Run("post remote dashboard - success", func(t *testing.T) {
		dashboard, err := os.ReadFile("testdata/test_json/post_dashboard.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(dashboard, &resource)
		require.NoError(t, err)

		err = handler.Add(resource)
		require.NoError(t, err)

		dash, err := handler.GetByUID("d4sHb0ard-")
		require.NoError(t, err)
		require.NotNil(t, dash)

		require.Equal(t, "abcdefghi", resource.GetMetadata("folder"))
	})

	t.Run("post remote dashboard - not found", func(t *testing.T) {
		resource := map[string]interface{}{
			"metadata": map[string]interface{}{
				"folder": "dummy",
				"name":   "dummy",
			},
		}

		err := handler.Add(resource)
		require.EqualError(t, err, "cannot upload dashboard dummy as folder dummy not found")
	})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			"metadata": map[string]interface{}{
				"name": "test",
			},
		}
		handler := DashboardHandler{}
		uid, err := handler.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, "test", uid)
	})

	_ = os.Unsetenv("GRAFANA_URL")
}

func TestExtractFolderUID(t *testing.T) {
	h := DashboardHandler{}
	os.Setenv("GRAFANA_URL", GetUrl())

	client, err := GetClient()
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
		uid := h.extractFolderUID(&mockFolderExtractor{}, dashboard)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid successfully - url exists", func(t *testing.T) {
		meta := models.DashboardMeta{
			FolderURL: "/dashboards/f/sample/special-sample-folder",
		}
		dashboard := models.DashboardFullWithMeta{
			Meta: &meta,
		}
		uid := h.extractFolderUID(&mockFolderExtractor{}, dashboard)
		require.Equal(t, "sample", uid)
	})

	t.Run("extract folder uid - empty uid returned", func(t *testing.T) {
		meta := models.DashboardMeta{
			FolderID: 1,
		}
		dashboard := models.DashboardFullWithMeta{
			Meta: &meta,
		}

		uid := h.extractFolderUID(&mockFolderExtractor{}, dashboard)
		require.Equal(t, "12345", uid)
	})
}

type mockFolderExtractor struct{}

func (m *mockFolderExtractor) getFolderById(folderId int64) (*models.Folder, error) {
	return &models.Folder{
		UID: "12345",
	}, nil
}
