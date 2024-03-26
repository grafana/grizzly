package integration_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestLibraryElements(t *testing.T) {
	provider, err := grafana.NewProvider(&testutil.TestContext().Grafana)
	require.NoError(t, err)
	handler := grafana.NewLibraryElementHandler(provider)

	t.Run("create libraryElement - success", func(t *testing.T) {
		libraryElement, err := os.ReadFile("testdata/test_json/post_library-element.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(libraryElement, &resource.Body)
		require.NoError(t, err)

		err = handler.Add(resource)
		require.NoError(t, err)

		remoteLibraryElement, err := handler.GetByUID("example-panel")
		require.NoError(t, err)
		require.NotNil(t, remoteLibraryElement)
		require.Equal(t, remoteLibraryElement.GetSpecValue("name").(string), "Example Panel")
	})

	t.Run("get remote libraryElement - success", func(t *testing.T) {
		resource, err := handler.GetByUID("example-panel")
		require.NoError(t, err)

		require.Equal(t, "grizzly.grafana.com/v1alpha1", resource.APIVersion())
		require.Equal(t, "example-panel", resource.Name())
		require.Len(t, resource.Spec(), 9)
	})

	t.Run("get remote libraryElement - not found", func(t *testing.T) {
		_, err := handler.GetByUID("dummy")
		require.ErrorContains(t, err, "Error retrieving library element dummy: not found")
	})

	t.Run("get libraryElements list", func(t *testing.T) {
		resources, err := handler.ListRemote()
		require.NoError(t, err)

		require.NotNil(t, resources)
		require.Len(t, resources, 1)
	})
}

func TestLibraryElements_CLI(t *testing.T) {
	dir := "testdata/library-elements"
	setupContexts(t, dir)

	t.Run("Apply library element in folder", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:                "apply in-folder/folder.json",
					ExpectedOutputContains: "DashboardFolder.xVvlWeK4k added",
				},
				{
					Command:                "apply in-folder/panel.json",
					ExpectedOutputContains: "LibraryElement.-agbm1C4k added",
				},
				// Test update
				{
					Command:                "apply in-folder/panel.json",
					ExpectedOutputContains: "LibraryElement.-agbm1C4k updated",
				},
				// Test list remote
				{
					Command:                "list -r -t LibraryElement",
					ExpectedOutputContains: "grizzly.grafana.com/v1alpha1    LibraryElement    -agbm1C4k",
				},
			},
		})
	})
}
