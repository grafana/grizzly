package mimir

import (
	"errors"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var errCortextoolClient = errors.New("error coming from cortextool client")

func TestRules(t *testing.T) {
	cortexTool := &FakeCortexTool{}
	h := RuleHandler{
		BaseHandler: grizzly.NewBaseHandler(&Provider{}, "PrometheusRuleGroup", false),
		clientTool:  cortexTool,
	}
	t.Run("get remote rule group", func(t *testing.T) {
		cortexTool.mockResponse(t, true, nil)
		res, err := h.getRemoteRuleGroup("first_rules.grizzly_alerts")
		require.NoError(t, err)
		uid, err := h.GetUID(*res)
		require.NoError(t, err)
		require.Equal(t, "grizzly_alerts", res.Name())
		require.Equal(t, "first_rules.grizzly_alerts", uid)
		require.Equal(t, "first_rules", res.GetMetadata("namespace"))
		require.Equal(t, "PrometheusRuleGroup", res.Kind())
	})

	t.Run("get remote rule group - error from cortextool client", func(t *testing.T) {
		cortexTool.mockResponse(t, false, errCortextoolClient)
		res, err := h.getRemoteRuleGroup("first_rules.grizzly_alerts")
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("get remote rule group - return not found", func(t *testing.T) {
		cortexTool.mockResponse(t, true, nil)
		res, err := h.getRemoteRuleGroup("name.name")
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("get remote rule group list", func(t *testing.T) {
		cortexTool.mockResponse(t, true, nil)
		res, err := h.getRemoteRuleGroupList()
		require.NoError(t, err)
		require.Equal(t, "first_rules.grizzly_alerts", res[0])
	})

	t.Run("get remote rule group list", func(t *testing.T) {
		cortexTool.mockResponse(t, false, errCortextoolClient)
		res, err := h.getRemoteRuleGroupList()
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("write rule group", func(t *testing.T) {
		cortexTool.mockResponse(t, false, nil)
		spec := make(map[string]interface{})
		file, err := os.ReadFile("testdata/rules.yaml")
		require.NoError(t, err)
		err = yaml.Unmarshal(file, &spec)
		require.NoError(t, err)

		resource, _ := grizzly.NewResource("apiV", "kind", "name", spec)
		resource.SetMetadata("namespace", "value")
		err = h.writeRuleGroup(resource)
		require.NoError(t, err)
	})

	t.Run("write rule group - error from the cortextool client", func(t *testing.T) {
		cortexTool.mockResponse(t, false, errCortextoolClient)
		spec := make(map[string]interface{})
		file, err := os.ReadFile("testdata/rules.yaml")
		require.NoError(t, err)
		err = yaml.Unmarshal(file, &spec)
		require.NoError(t, err)

		resource, _ := grizzly.NewResource("apiV", "kind", "name", spec)
		resource.SetMetadata("namespace", "value")
		err = h.writeRuleGroup(resource)
		require.Error(t, err)
	})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			Body: map[string]any{
				"metadata": map[string]any{
					"name":      "test",
					"namespace": "test_namespace",
				},
			},
		}
		uid, err := h.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, "test_namespace.test", uid)
	})
}

type FakeCortexTool struct {
	hasFile       bool
	expectedError error
}

func (f *FakeCortexTool) ListRules() (map[string][]models.PrometheusRuleGroup, error) {
	if f.expectedError != nil {
		return nil, f.expectedError
	}

	if f.hasFile {
		res, err := os.ReadFile("testdata/list_rules.yaml")
		if err != nil {
			return nil, err
		}

		var group map[string][]models.PrometheusRuleGroup
		if err := yaml.Unmarshal(res, &group); err != nil {
			return nil, err
		}

		return group, nil
	}

	return nil, nil
}

func (f *FakeCortexTool) LoadRules(_ models.PrometheusRuleGrouping) (string, error) {
	if f.expectedError != nil {
		return "", f.expectedError
	}

	if f.hasFile {
		res, err := os.ReadFile("testdata/list_rules.yaml")
		if err != nil {
			return "", err
		}

		return string(res), nil
	}

	return "", nil
}

func (f *FakeCortexTool) mockResponse(t *testing.T, hasFile bool, expectedError error) {
	f.hasFile = hasFile
	f.expectedError = expectedError
	t.Cleanup(func() {
		f.hasFile = false
		f.expectedError = nil
	})
}
