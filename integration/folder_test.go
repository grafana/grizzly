package integration_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	. "github.com/grafana/grizzly/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestFolders(t *testing.T) {
	InitialiseTestConfig()
	handler := grafana.NewFolderHandler(grafana.NewProvider())

	t.Run("get remote folder - success", func(t *testing.T) {
		resource, err := handler.GetByUID("abcdefghi")
		require.NoError(t, err)

		require.Equal(t, "grizzly.grafana.com/v1alpha1", resource.APIVersion())
		require.Equal(t, "abcdefghi", resource.Name())
		require.Len(t, resource.Spec(), 14)
	})

	t.Run("get remote folder - not found", func(t *testing.T) {
		_, err := handler.GetByUID("dummy")
		require.ErrorContains(t, err, "Couldn't fetch folder 'dummy' from remote: not found")
	})

	t.Run("get folders list", func(t *testing.T) {
		resources, err := handler.ListRemote()
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

		err = handler.Add(resource)
		require.NoError(t, err)

		remoteFolder, err := handler.GetByUID("newFolder")
		require.NoError(t, err)
		require.NotNil(t, remoteFolder)
		require.Equal(t, "/dashboards/f/newFolder/new-folder", remoteFolder.Spec()["url"])

		t.Run("conflict: put remote folder - update uid", func(t *testing.T) {
			remoteFolder.SetSpecString("uid", "dummyUid")

			err := handler.Add(*remoteFolder)
			apiError := err.(grafana.APIResponse)
			require.Equal(t, 409, apiError.Code())
		})
	})

	t.Run("post remote folder - conflict - folder already exists", func(t *testing.T) {
		folder, err := os.ReadFile("testdata/test_json/post_folder.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(folder, &resource)
		require.NoError(t, err)

		resource.SetSpecString("title", "Azure Data Explorer")

		err = handler.Add(resource)
		apiError := err.(*runtime.APIError)
		require.Equal(t, 412, apiError.Code)
	})
}
