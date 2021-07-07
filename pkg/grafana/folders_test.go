package grafana

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestFolders(t *testing.T) {
	os.Setenv("GRAFANA_URL", "http://grizzly-grafana:3000")

	// ctx := context.Background()
	// cli, err := initClient(ctx)
	// require.NoError(t, err)

	// containerID := startContainer(cli, ctx)

	// go func() {
	// 	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	// 	select {
	// 	case err := <-errCh:
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	case <-statusCh:
	// 	}
	// }()

	ticker := pingLocalhost()
	defer ticker.Stop()

	// defer func() {
	// 	removeContainer(cli, ctx, containerID)
	// }()

	// printContainerLogs(cli, ctx, containerID)

	t.Run("get remote folder - success", func(t *testing.T) {
		resource, err := getRemoteFolder("abcdefghi")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "abcdefghi")
		require.Len(t, resource.Spec(), 13)
	})

	t.Run("get remote folder - not found", func(t *testing.T) {
		_, err := getRemoteFolder("dummy")
		require.EqualError(t, err, "couldn't fetch folder 'dummy' from remote: couldn't fetch folder 'dummy' from remote: not found")
	})

	t.Run("get folders list", func(t *testing.T) {
		resources, err := getRemoteFolderList()
		require.NoError(t, err)

		require.NotNil(t, resources)
		require.Len(t, resources, 1)
	})

	t.Run("post remote folder - success", func(t *testing.T) {
		folder, err := os.ReadFile("testdata/test_json/post_folder.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(folder, &resource)
		require.NoError(t, err)

		err = postFolder(resource)
		require.NoError(t, err)

		remoteFolder, err := getRemoteFolder("newFolder")
		require.NoError(t, err)
		require.NotNil(t, remoteFolder)
		require.Equal(t, remoteFolder.Spec()["url"], "/dashboards/f/newFolder/new-folder")

		t.Run("put remote folder - update uid", func(t *testing.T) {
			remoteFolder.SetSpecString("uid", "dummyUid")

			err := putFolder(*remoteFolder)
			require.NoError(t, err)

			updatedFolder, err := getRemoteFolder("dummyUid")
			require.NoError(t, err)

			require.Equal(t, updatedFolder.Spec()["uid"], "dummyUid")
		})
	})

	t.Run("post remote folder - conflict - folder already exists", func(t *testing.T) {
		folder, err := os.ReadFile("testdata/test_json/post_folder.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(folder, &resource)
		require.NoError(t, err)

		resource.SetSpecString("title", "Azure Data Explorer")

		err = postFolder(resource)

		grafanaErr := err.(ErrNon200Response)
		require.Error(t, err)
		// TODO: change to 409 after this PR is in the new Grafana version https://github.com/grafana/grafana/pull/36429
		require.Equal(t, grafanaErr.Response.StatusCode, 400)
	})
}
