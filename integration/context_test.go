package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContexts(t *testing.T) {
	dir := "testdata/contexts"
	setupContexts(t, dir)

	t.Run("Get contexts - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: dir,
			Commands: []Command{
				{
					Command:            "config get-contexts",
					ExpectedCode:       0,
					ExpectedOutputFile: "get-contexts.txt",
				},
			},
		})
	})

	absConfigPath, err := filepath.Abs("testdata/contexts/settings.yaml")
	require.NoError(t, err)
	t.Run("Find config path", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: dir,
			Commands: []Command{
				{
					Command:        "config path",
					ExpectedOutput: absConfigPath,
				},
			},
		})
	})

	t.Run("Get context config", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: dir,
			Commands: []Command{
				// Whole config
				{
					Command:            "config get",
					ExpectedOutputFile: "get-context-val.yml",
				},
				// Whole config JSON
				{
					Command:            "config get -o json",
					ExpectedOutputFile: "get-context-val.json",
				},
				// Specific key
				{
					Command:        "config get grafana.url",
					ExpectedOutput: "http://localhost:3001",
				},
			},
		})
	})

	t.Run("Get SM URL", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: "testdata/contexts",
			Commands: []Command{
				{
					Command:        "config get synthetic-monitoring.api-url",
					ExpectedOutput: "https://synthetic-monitoring-api.grafana.net",
				},
				{
					Command:        "config get synthetic-monitoring.api-url",
					ExpectedOutput: "https://custom-url.com",
					Env: map[string]string{
						"GRAFANA_SM_API_URL": "https://custom-url.com",
					},
				},
			},
		})
	})
}
