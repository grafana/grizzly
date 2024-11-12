package grafana

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestAlertRuleGroupHandler_ResourceFilePath(t *testing.T) {
	handler := NewAlertRuleGroupHandler(&Provider{})

	t.Run("slashes are escaped from filenames", func(t *testing.T) {
		req := require.New(t)

		resource, err := grizzly.NewResource(handler.APIVersion(), handler.Kind(), "some/alert/group", map[string]interface{}{})
		req.NoError(err)

		req.Equal("alert-rules/alertRuleGroup-some-alert-group.yaml", handler.ResourceFilePath(resource, "yaml"))
	})
}
