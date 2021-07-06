package grafana

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types/container"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDatasources(t *testing.T) {
	os.Setenv("GRAFANA_URL", "tcp://0.0.0.0:3000")

	ctx := context.Background()
	cli, err := initClient(ctx)
	require.NoError(t, err)

	containerID := startContainer(err, cli, ctx)

	go func() {
		statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
		}
	}()

	ticker := pingLocalhost()
	defer ticker.Stop()

	defer func() {
		removeContainer(cli, ctx, containerID)
	}()

	printContainerLogs(cli, ctx, containerID)
	t.Run("get remote datasource - success", func(t *testing.T) {
		resource, err := getRemoteDatasource("AppDynamics")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "AppDynamics")
		require.Len(t, resource.Spec(), 20)
	})

	t.Run("get remote datasource - not found", func(t *testing.T) {
		_, err := getRemoteDatasource("dummy")
		require.EqualError(t, err, "couldn't fetch folder 'dummy' from remote: not found")
	})

	t.Run("get remote datasources list", func(t *testing.T) {
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

		ds, err := getRemoteDatasource("appdynamics")
		require.NoError(t, err)
		require.NotNil(t, ds)
		require.Equal(t, ds.Spec()["type"], "dlopes7-appdynamics-datasource")

		t.Run("put remote datasource - update", func(t *testing.T) {
			ds.SetSpecString("type", "new-type")

			err := putDatasource(*ds)
			require.NoError(t, err)

			updatedDS, err := getRemoteDatasource("appdynamics")
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

		grafanaErr := err.(ErrNon200Response)
		require.Error(t, err)
		require.Equal(t, grafanaErr.Response.StatusCode, 409)
	})
}
