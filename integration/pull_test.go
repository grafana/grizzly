package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPull(t *testing.T) {
	dir := t.TempDir()
	setupContexts(t, dir)

	t.Run("Pull everything - success", func(t *testing.T) {
		pullDir := t.TempDir()
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:                fmt.Sprintf("pull %s -t Dashboard -t Dashboardfolder -t Datasource", pullDir),
					ExpectedCode:           0,
					ExpectedOutputContains: `Dashboard.ReciqtgGk pulled`,
				},
			},
			Validate: func(t *testing.T) {
				// Check the files
				assert.DirExists(t, filepath.Join(pullDir, "dashboards"))
				assert.FileExists(t, filepath.Join(pullDir, "dashboards", "abcdefghi", "dashboard-ReciqtgGk.yaml"))
				assert.DirExists(t, filepath.Join(pullDir, "datasources"))
				assert.DirExists(t, filepath.Join(pullDir, "folders"))
			},
		})
	})
}
