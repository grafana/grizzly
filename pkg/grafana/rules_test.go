package grafana

import (
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRules(t *testing.T) {
	t.Run("get remote rule group", func(t *testing.T) {
		file, err := os.ReadFile("testdata/rules.yaml")
		require.NoError(t, err)
		client = mockCTClient{
			rules: file,
		}
		_, err = getRemoteRuleGroup("rules.rules")
		require.NoError(t, err)
	})

	t.Run("get remote rule group list", func(t *testing.T) {
		file, err := os.ReadFile("testdata/rules.yaml")
		require.NoError(t, err)
		client = mockCTClient{
			rules: file,
		}
		_, err = getRemoteRuleGroupList()
		require.NoError(t, err)
	})

	t.Run("write rule group", func(t *testing.T) {
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
