package integration_test

import (
	"testing"
)

func TestDashboard(t *testing.T) {
	dir := "testdata/dashboards"
	setupContexts(t, dir)

	t.Run("Get dashboard - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:            "get Dashboard.ReciqtgGk",
					ExpectedCode:       0,
					ExpectedOutputFile: "ReciqtgGk.json",
				},
			},
		})
	})

	t.Run("Get dashboard list - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:            "list -r -t Dashboard",
					ExpectedOutputFile: "list.txt",
				},
			},
		})
	})

	t.Run("Apply dashboard - no folder", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:        "apply no-folder.yml",
					ExpectedOutput: "Dashboard.no-folder added\n",
				},
				{
					Command:                "get Dashboard.no-folder",
					ExpectedOutputContains: "folder: general",
				},
			},
		})
	})

	t.Run("Get dashboard - failure", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:      "get missing-dashboard",
					ExpectedCode: 1,
				},
			},
		})
	})
}
