package grafana

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestSyntheticMonitoring(t *testing.T) {

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{},
		})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			"metadata": map[string]interface{}{
				"name": "test",
				"type": "http",
			},
		}
		handler := SyntheticMonitoringHandler{}
		uid, err := handler.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, uid, "http.test")
	})
}
