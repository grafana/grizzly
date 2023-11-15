package grafana

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	. "github.com/grafana/grizzly/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestDatasources(t *testing.T) {
	os.Setenv("GRAFANA_URL", GetUrl())

	grafanaClient, err := GetClient()
	require.NoError(t, err)

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			NewProvider(grafanaClient),
		})

	ticker := PingService(GetUrl())
	defer ticker.Stop()

	handler, err := grizzly.Registry.GetHandler((&DatasourceHandler{}).Kind())
	require.NoError(t, err)

	t.Run("get remote datasource - success", func(t *testing.T) {
		resource, err := handler.GetByUID("AppDynamics")
		require.NoError(t, err)

		require.Equal(t, "grizzly.grafana.com/v1alpha1", resource.APIVersion())
		require.Equal(t, "AppDynamics", resource.Name())
		require.Len(t, resource.Spec(), 12)
	})

	t.Run("get remote datasource - not found", func(t *testing.T) {
		_, err := handler.GetByUID("dummy")
		require.Equal(t, grizzly.ErrNotFound, err)
	})

	t.Run("get remote datasources list", func(t *testing.T) {
		resources, err := handler.ListRemote()
		require.NoError(t, err)

		require.NotNil(t, resources)
		require.Len(t, resources, 1)
	})

	t.Run("post remote datasource - success", func(t *testing.T) {
		datasource, err := os.ReadFile("testdata/test_json/post_datasource.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(datasource, &resource)
		require.NoError(t, err)

		err = handler.Add(resource)
		require.NoError(t, err)

		ds, err := handler.GetByUID("appdynamics")
		require.NoError(t, err)
		require.NotNil(t, ds)
		require.Equal(t, "dlopes7-appdynamics-datasource", ds.Spec()["type"])

		t.Run("put remote datasource - update", func(t *testing.T) {
			ds.SetSpecString("type", "new-type")

			err := handler.Update(nil, *ds)
			require.NoError(t, err)

			updatedDS, err := handler.GetByUID("appdynamics")
			require.NoError(t, err)

			require.Equal(t, "new-type", updatedDS.Spec()["type"])
		})
	})

	t.Run("post remote datasource - conflict - datasource already exists", func(t *testing.T) {
		datasource, err := os.ReadFile("testdata/test_json/post_datasource.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(datasource, &resource)
		require.NoError(t, err)

		resource.SetSpecString("name", "AppDynamics")

		err = handler.Add(resource)

		var non200ResponseErr ErrNon200Response
		require.ErrorAs(t, err, &non200ResponseErr)
		require.Equal(t, 409, non200ResponseErr.Response.StatusCode)
	})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			"metadata": map[string]interface{}{
				"name": "test",
			},
		}
		handler := DatasourceHandler{}
		uid, err := handler.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, "test", uid)
	})
}
