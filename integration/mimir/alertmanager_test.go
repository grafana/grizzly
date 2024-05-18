package mimir

import (
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir"
	"github.com/grafana/grizzly/pkg/mimir/client"
	"github.com/grafana/grizzly/pkg/testutil"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAlertmanager(t *testing.T) {
	alertmanagerTestFilePath := "testdata/alertmanager/test-alertmanager.yml"
	provider := mimir.NewProvider(&testutil.TestContext().Mimir)
	client := client.NewHTTPClient(&testutil.TestContext().Mimir)
	handler := mimir.NewAlertmanagerHandler(provider, client)

	t.Run("create prometheus alertmanager config", func(t *testing.T) {

		file, err := os.ReadFile(alertmanagerTestFilePath)
		require.NoError(t, err)

		var resource grizzly.Resource
		require.NoError(t, yaml.Unmarshal(file, &resource.Body))
		require.NoError(t, handler.Add(resource))

		t.Run("get remote alertmanager config", func(t *testing.T) {
			file, err := os.ReadFile(alertmanagerTestFilePath)
			require.NoError(t, err)

			var resource grizzly.Resource
			require.NoError(t, yaml.Unmarshal(file, &resource.Body))

			remoteResource, err := handler.GetRemote(grizzly.Resource{})
			require.NoError(t, err)

			require.Equal(t, resource, *remoteResource)
		})
	})
}
