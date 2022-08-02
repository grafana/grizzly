package grafana

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestAlert(t *testing.T) {
	os.Setenv("GRAFANA_URL", getUrl())

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{},
		})

	ticker := pingService(getUrl())
	defer ticker.Stop()

	t.Run("post alert folder - success", func(t *testing.T) {
		folder, err := os.ReadFile("testdata/test_json/post_alert_folder.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(folder, &resource)
		require.NoError(t, err)

		err = postFolder(resource)
		require.NoError(t, err)

		remoteFolder, err := getRemoteFolder("alertTestNamespace")
		require.NoError(t, err)
		require.NotNil(t, remoteFolder)
		require.Equal(t, remoteFolder.Spec()["url"], "/dashboards/f/alertTestNamespace/alerttestnamespace")
	})

	t.Run("post remote alert - success", func(t *testing.T) {
		alertJSON, err := os.ReadFile("testdata/test_json/post_alert.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(alertJSON, &resource)
		require.NoError(t, err)

		err = postAlertGroup(resource)
		require.NoError(t, err)

		alert, err := getRemoteAlertGroup("AlertTest/AlertTest")
		require.NoError(t, err)
		require.NotNil(t, alert)

		require.Equal(t, resource.GetMetadata("folder"), "abcdefghi")
	})

	// t.Run("get remote alert - success", func(t *testing.T) {
	// 	resource, err := getRemoteAlertGroup("grafana||first_alerts/grizzly_alerts")
	// 	require.NoError(t, err)

	// 	require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
	// 	require.Equal(t, resource.Name(), "ReciqtgGk")
	// 	require.NotEmpty(t, resource.GetMetadata("folder"))
	// })

	// t.Run("get remote alert - not found", func(t *testing.T) {
	// 	_, err := getRemoteAlertGroup("grafana||7/bc")
	// 	require.EqualError(t, err, "not found")
	// })

	// t.Run("get remote alert list - success", func(t *testing.T) {
	// 	list, err := getRemoteAlertGroupList()
	// 	require.NoError(t, err)

	// 	require.Len(t, list, 3)
	// 	require.EqualValues(t, []string{"ReciqtgGk", "392Ik4GGk", "kE0IIVGGz"}, list)
	// })

	// t.Run("post remote alert - not found", func(t *testing.T) {
	// 	var resource grizzly.Resource
	// 	resource = map[string]interface{}{
	// 		"metadata": map[string]interface{}{
	// 			"folder": "dummy",
	// 			"name":   "dummy",
	// 		},
	// 	}

	// 	err := postAlertGroup(resource)
	// 	require.EqualError(t, err, "Cannot upload alert dummy as folder dummy not found")
	// })

	// t.Run("Check getUID is functioning correctly", func(t *testing.T) {
	// 	resource := grizzly.Resource{
	// 		"metadata": map[string]interface{}{
	// 			"name": "test",
	// 		},
	// 	}
	// 	handler := AlertsHandler{}
	// 	uid, err := handler.GetUID(resource)
	// 	require.NoError(t, err)
	// 	require.Equal(t, uid, "test")
	// })

	_ = os.Unsetenv("GRAFANA_URL")
}
