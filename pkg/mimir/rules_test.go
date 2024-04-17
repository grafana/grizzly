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

var errMimirClient = errors.New("error coming from mimir client")

func TestRules(t *testing.T) {
	client := &FakeClient{}
	h := RuleHandler{
		BaseHandler: grizzly.NewBaseHandler(&Provider{}, "PrometheusRuleGroup", false),
		clientTool:  client,
	}
	t.Run("get remote rule group", func(t *testing.T) {
		client.mockResponse(t, true, nil)
		res, err := h.getRemoteRuleGroup("first_rules.grizzly_alerts")
		require.NoError(t, err)
		uid, err := h.GetUID(*res)
		require.NoError(t, err)
		require.Equal(t, "grizzly_alerts", res.Name())
		require.Equal(t, "first_rules.grizzly_alerts", uid)
		require.Equal(t, "first_rules", res.GetMetadata("namespace"))
		require.Equal(t, "PrometheusRuleGroup", res.Kind())
	})
	
	t.Run("get remote rule group - error from mimir client", func(t *testing.T) {
		client.mockResponse(t, false, errMimirClient)
		res, err := h.getRemoteRuleGroup("first_rules.grizzly_alerts")
		require.Error(t, err)
		require.Nil(t, res)
	})
	
	t.Run("get remote rule group - return not found", func(t *testing.T) {
		client.mockResponse(t, true, nil)
		res, err := h.getRemoteRuleGroup("name.name")
		require.Error(t, err)
		require.Nil(t, res)
	})
	
	t.Run("get remote rule group list", func(t *testing.T) {
		client.mockResponse(t, true, nil)
		res, err := h.getRemoteRuleGroupList()
		require.NoError(t, err)
		require.Equal(t, "first_rules.grizzly_alerts", res[0])
	})
	
	t.Run("get remote rule group list", func(t *testing.T) {
		client.mockResponse(t, false, errMimirClient)
		res, err := h.getRemoteRuleGroupList()
		require.Error(t, err)
		require.Nil(t, res)
	})
	
	t.Run("write rule group", func(t *testing.T) {
		client.mockResponse(t, false, nil)
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
	
	t.Run("write rule group - error from the mimir client", func(t *testing.T) {
		client.mockResponse(t, false, errMimirClient)
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

type FakeClient struct {
	hasFile       bool
	expectedError error
}

func (f *FakeClient) ListRules() (map[string][]models.PrometheusRuleGroup, error) {
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

func (f *FakeClient) CreateRules(_ models.PrometheusRuleGrouping) error {
	if f.expectedError != nil {
		return f.expectedError
	}
	
	if f.hasFile {
		_, err := os.ReadFile("testdata/list_rules.yaml")
		if err != nil {
			return err
		}
		
		return nil
	}
	
	return nil
}

func (f *FakeClient) mockResponse(t *testing.T, hasFile bool, expectedError error) {
	f.hasFile = hasFile
	f.expectedError = expectedError
	t.Cleanup(func() {
		f.hasFile = false
		f.expectedError = nil
	})
}
