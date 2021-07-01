package grafana

import (
	"errors"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDashboards(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	_ = os.Setenv("GRAFANA_URL", "url")
	_ = os.Setenv("GRAFANA_TOKEN", "token")
	_ = os.Setenv("GRAFANA_USER", "user")

	t.Run("valid dashboard response", func(t *testing.T) {
		expected, err := os.ReadFile("testdata/dashboards/valid_dashboard.json")
		require.NoError(t, err)

		httpmock.RegisterResponder("GET", "//user:token@url/api/dashboards/uid/uid",
			httpmock.NewStringResponder(200, string(expected)))
		resource, err := getRemoteDashboard("uid")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "uid")
		require.Equal(t, resource.GetMetadata("folder"), "sample")
	})

	t.Run("invalid response from http call", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "//user:token@url/api/dashboards/uid/uid",
			httpmock.NewStringResponder(200, "abc123"))
		_, err := getRemoteDashboard("uid")
		require.Error(t, err)

		var apiErr grizzly.APIErr
		require.True(t, errors.As(err, &apiErr))
	})

	t.Run("valid dashboard response - folderUid doesn't exist", func(t *testing.T) {
		expectedDashboard, err := os.ReadFile("testdata/dashboards/valid_dashboard.json")
		expectedFolder := `{ "uid": 4 }`
		require.NoError(t, err)

		httpmock.RegisterResponder("GET", "//user:token@url/api/dashboards/uid/uid",
			httpmock.NewStringResponder(200, string(expectedDashboard)))
		httpmock.RegisterResponder("GET", "//user:token@url/api/folder/id/4",
			httpmock.NewStringResponder(200, expectedFolder))
		resource, err := getRemoteDashboard("uid")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "uid")
		require.Equal(t, resource.GetMetadata("folder"), "sample")
	})

	_ = os.Unsetenv("GRAFANA_URL")
	_ = os.Unsetenv("GRAFANA_TOKEN")
	_ = os.Unsetenv("GRAFANA_USER")
}
