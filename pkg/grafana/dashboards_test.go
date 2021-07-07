package grafana

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestDashboard(t *testing.T) {
	os.Setenv("GRAFANA_URL", "http://localhost:3000")

	ctx := context.Background()
	cli, err := initClient(ctx)
	require.NoError(t, err)

	containerID := startContainer(cli, ctx)

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

	ticker := pingLocalhost(cli, ctx, containerID)
	defer ticker.Stop()

	defer func() {
		removeContainer(cli, ctx, containerID)
	}()

	printContainerLogs(cli, ctx, containerID)

	t.Run("get remote dashboard - success", func(t *testing.T) {
		resource, err := getRemoteDashboard("ReciqtgGk")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "ReciqtgGk")
		require.NotEmpty(t, resource.GetMetadata("folder"))
	})

	t.Run("get remote dashboard - not found", func(t *testing.T) {
		_, err := getRemoteDashboard("dummy")
		require.EqualError(t, err, "not found")
	})

	t.Run("get remote dashboard list - success", func(t *testing.T) {
		list, err := getRemoteDashboardList()
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

		err = postDashboard(resource)
		require.NoError(t, err)

		dash, err := getRemoteDashboard("d4sHb0ard-")
		require.NoError(t, err)
		require.NotNil(t, dash)

		require.Equal(t, resource.GetMetadata("folder"), "abcdefghi")
	})

	t.Run("post remote dashboard - not found", func(t *testing.T) {
		var resource grizzly.Resource
		resource = map[string]interface{}{
			"metadata": map[string]interface{}{
				"folder": "dummy",
				"name":   "dummy",
			},
		}

		err := postDashboard(resource)
		require.EqualError(t, err, "Cannot upload dashboard dummy as folder dummy not found")
	})

	_ = os.Unsetenv("GRAFANA_URL")
}
