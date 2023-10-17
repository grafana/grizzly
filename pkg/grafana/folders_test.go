package grafana

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	. "github.com/grafana/grizzly/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestFolders(t *testing.T) {
	os.Setenv("GRAFANA_URL", GetUrl())

	client, err := GetClient()
	require.NoError(t, err)

	ticker := PingService(GetUrl())
	defer ticker.Stop()

	t.Run("get remote folder - success", func(t *testing.T) {
		resource, err := getRemoteFolder(client, "abcdefghi")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "abcdefghi")
		require.Len(t, resource.Spec(), 14)
	})

	t.Run("get remote folder - not found", func(t *testing.T) {
		_, err := getRemoteFolder(client, "dummy")
		require.EqualError(t, err, "couldn't fetch folder 'dummy' from remote: not found")
	})

	t.Run("get folders list", func(t *testing.T) {
		resources, err := getRemoteFolderList(client)
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

		err = postFolder(client, resource)
		require.NoError(t, err)

		remoteFolder, err := getRemoteFolder(client, "newFolder")
		require.NoError(t, err)
		require.NotNil(t, remoteFolder)
		require.Equal(t, remoteFolder.Spec()["url"], "/dashboards/f/newFolder/new-folder")

		t.Run("put remote folder - update uid", func(t *testing.T) {
			remoteFolder.SetSpecString("uid", "dummyUid")

			err := putFolder(client, *remoteFolder)
			require.NoError(t, err)

			updatedFolder, err := getRemoteFolder(client, "dummyUid")
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

		err = postFolder(client, resource)

		grafanaErr := err.(ErrNon200Response)
		require.Error(t, err)
		require.Equal(t, grafanaErr.Response.StatusCode, 409)
	})
}
