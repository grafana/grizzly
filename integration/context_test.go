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

	t.Run("Unset value", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: dir,
			Commands: []Command{
				{
					Command:             "config set mimir.invalid http://mimir:9009",
					ExpectedLogsContain: "key not recognised: mimir.invalid",
					ExpectedCode:        1,
				},
				{
					Command: "config set mimir.address http://mimir:9009",
				},
				{
					Command: "config set mimir.tenant-id tenant-id",
				},
				{
					Command:        "config get mimir -o yaml",
					ExpectedOutput: "address: http://mimir:9009\ntenant-id: tenant-id",
				},
				{
					Command: "config unset mimir.address",
				},
				{
					Command:        "config get mimir -o yaml",
					ExpectedOutput: "tenant-id: tenant-id",
				},
				{
					Command: "config unset mimir.tenant-id",
				},
				{
					Command:             "config get mimir -o yaml",
					ExpectedLogsContain: "key not found: mimir",
					ExpectedCode:        1,
				},
				{
					Command:             "config unset mimir.invalid",
					ExpectedLogsContain: "mimir.invalid is not a valid path",
					ExpectedCode:        1,
				},
				{
					Command:             "config unset mimir.tenant-id",
					ExpectedLogsContain: "key mimir.tenant-id is already unset",
					ExpectedCode:        1,
				},
			},
		})
	})
}
