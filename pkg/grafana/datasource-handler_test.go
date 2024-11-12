package grafana

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestDatasourceHandler_ResourceFilePath(t *testing.T) {
	handler := NewDatasourceHandler(&Provider{})

	t.Run("slashes are escaped from filenames", func(t *testing.T) {
		req := require.New(t)

		resource, err := grizzly.NewResource(handler.APIVersion(), handler.Kind(), "some/datasource", map[string]interface{}{})
		req.NoError(err)

		req.Equal("datasources/datasource-some-datasource.yaml", handler.ResourceFilePath(resource, "yaml"))
	})
}
