package grafana

import (
	"errors"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var errCortextoolClient = errors.New("error coming from cortextool client")

func TestRules(t *testing.T) {

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{},
		})

	t.Run("get remote rule group", func(t *testing.T) {
		mockCortexTool(t, "testdata/list_rules.yaml", nil)

		res, err := getRemoteRuleGroup("first_rules.grizzly_alerts")
		require.NoError(t, err)
		handler := RuleHandler{}
		uid, err := handler.GetUID(*res)
		require.NoError(t, err)
		require.Equal(t, "grizzly_alerts", res.Name())
		require.Equal(t, "first_rules.grizzly_alerts", uid)
		require.Equal(t, "first_rules", res.GetMetadata("namespace"))
		require.Equal(t, "PrometheusRuleGroup", res.Kind())
		key := res.Key()
		require.NoError(t, err)
		require.Equal(t, "PrometheusRuleGroup.first_rules.grizzly_alerts", key)
	})

	t.Run("get remote rule group - error from cortextool client", func(t *testing.T) {
		mockCortexTool(t, "", errCortextoolClient)

		res, err := getRemoteRuleGroup("first_rules.grizzly_alerts")
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("get remote rule group - return not found", func(t *testing.T) {
		mockCortexTool(t, "testdata/list_rules.yaml", nil)

		res, err := getRemoteRuleGroup("name.name")
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("get remote rule group list", func(t *testing.T) {
		mockCortexTool(t, "testdata/list_rules.yaml", nil)

		res, err := getRemoteRuleGroupList()
		require.NoError(t, err)
		require.Equal(t, "first_rules.grizzly_alerts", res[0])
	})

	t.Run("get remote rule group list", func(t *testing.T) {
		mockCortexTool(t, "", errCortextoolClient)

		res, err := getRemoteRuleGroupList()
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("write rule group", func(t *testing.T) {
		mockCortexTool(t, "", nil)

		spec := make(map[string]interface{})
		file, err := os.ReadFile("testdata/rules.yaml")
		require.NoError(t, err)
		err = yaml.Unmarshal(file, &spec)
		require.NoError(t, err)

		resource := grizzly.NewResource("apiV", "kind", "name", spec)
		resource.SetMetadata("namespace", "value")
		err = writeRuleGroup(resource)
		require.NoError(t, err)
	})

	t.Run("write rule group - error from the cortextool client", func(t *testing.T) {
		mockCortexTool(t, "", errCortextoolClient)

		spec := make(map[string]interface{})
		file, err := os.ReadFile("testdata/rules.yaml")
		require.NoError(t, err)
		err = yaml.Unmarshal(file, &spec)
		require.NoError(t, err)

		resource := grizzly.NewResource("apiV", "kind", "name", spec)
		resource.SetMetadata("namespace", "value")
		err = writeRuleGroup(resource)
		require.Error(t, err)
	})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "test_namespace",
			},
		}
		handler := RuleHandler{}
		uid, err := handler.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, uid, "test_namespace.test")
	})
}

func mockCortexTool(t *testing.T, file string, err error) {
	origCorexTool := cortexTool
	cortexTool = func(args ...string) ([]byte, error) {
		if file != "" {
			bytes, errFile := os.ReadFile("testdata/list_rules.yaml")
			require.NoError(t, errFile)

			return bytes, nil
		}

		return nil, err
	}
	t.Cleanup(func() {
		cortexTool = origCorexTool
	})
}

type mockCTClient struct {
	rules []byte
	err   error
}

func (m mockCTClient) listRules() ([]byte, error) {
	return m.rules, m.err
}

func (m mockCTClient) writeRules(namespace, fileName string) error {
	return m.err
}
