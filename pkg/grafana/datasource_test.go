package grafana

import (
	"encoding/json"
	gapi "github.com/grafana/grafana-api-golang-client"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestDataSources(t *testing.T) {
	os.Setenv("GRAFANA_URL", getUrl())

	client, err := gapi.New(getUrl(), gapi.Config{})
	require.NoError(t, err)

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{client: client},
		})

	ticker := pingService(getUrl())
	defer ticker.Stop()

	handler, err := grizzly.Registry.GetHandler((&DatasourceHandler{}).Kind())
	require.NoError(t, err)

	t.Run("get remote datasource - success", func(t *testing.T) {
		resource, err := handler.GetByUID("AppDynamics")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "AppDynamics")
		require.Len(t, resource.Spec(), 18)
	})

	t.Run("get remote datasource - not found", func(t *testing.T) {
		_, err := handler.GetByUID("dummy")
		require.Equal(t, err, grizzly.ErrNotFound)
	})

	t.Run("get remote data sources list", func(t *testing.T) {
		resources, err := getRemoteDatasourceList()
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

		err = postDatasource(resource)
		require.NoError(t, err)

		ds, err := handler.GetByUID("appdynamics")
		require.NoError(t, err)
		require.NotNil(t, ds)
		require.Equal(t, ds.Spec()["type"], "dlopes7-appdynamics-datasource")

		t.Run("put remote datasource - update", func(t *testing.T) {
			ds.SetSpecString("type", "new-type")

			err := putDatasource(*ds)
			require.NoError(t, err)

			updatedDS, err := handler.GetByUID("appdynamics")
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

		err = postDatasource(resource)

		var non200ResponseErr ErrNon200Response
		require.ErrorAs(t, err, &non200ResponseErr)
		require.Equal(t, non200ResponseErr.Response.StatusCode, 409)
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
