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

	client, err := GetClient()
	require.NoError(t, err)

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{client: client},
		})

	ticker := PingService(GetUrl())
	defer ticker.Stop()

	t.Run("get remote datasource - success", func(t *testing.T) {
		resource, err := getRemoteDatasource(client, "AppDynamics")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "AppDynamics")
		require.Len(t, resource.Spec(), 18)
	})

	t.Run("get remote datasource - not found", func(t *testing.T) {
		_, err := getRemoteDatasource(client, "dummy")
		require.Equal(t, err, grizzly.ErrNotFound)
	})

	t.Run("get remote datasources list", func(t *testing.T) {
		resources, err := getRemoteDatasourceList(client)
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

		err = postDatasource(client, resource)
		require.NoError(t, err)

		ds, err := getRemoteDatasource(client, "appdynamics")
		require.NoError(t, err)
		require.NotNil(t, ds)
		require.Equal(t, ds.Spec()["type"], "dlopes7-appdynamics-datasource")

		t.Run("put remote datasource - update", func(t *testing.T) {
			ds.SetSpecString("type", "new-type")

			err := putDatasource(client, *ds)
			require.NoError(t, err)

			updatedDS, err := getRemoteDatasource(client, "appdynamics")
			require.NoError(t, err)

			require.Equal(t, updatedDS.Spec()["type"], "new-type")
		})
	})

	t.Run("post remote datasource - conflict - datasource already exists", func(t *testing.T) {
		datasource, err := os.ReadFile("testdata/test_json/post_datasource.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(datasource, &resource)
		require.NoError(t, err)

		resource.SetSpecString("name", "AppDynamics")

		err = postDatasource(client, resource)

		grafanaErr := err.(ErrNon200Response)
		require.Error(t, err)
		require.Equal(t, grafanaErr.Response.StatusCode, 409)
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
		require.Equal(t, uid, "test")
	})
}
