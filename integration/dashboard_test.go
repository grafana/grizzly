package integration_test

import (
	"testing"

	"github.com/grafana/grizzly/pkg/testutil"
)

func TestDashboard(t *testing.T) {
	ticker := testutil.PingService(testutil.GetUrl())
	defer ticker.Stop()
	dir := "testdata/dashboards"
	setupContexts(t, dir)

	tests := []GrizzlyTest{
		{
			Name:    "Get dashboard - success",
			TestDir: dir,
			Commands: []Command{
				{
					Command:        "get Dashboard.ReciqtgGk",
					ExpectedCode:   0,
					ExpectedOutput: "ReciqtgGk.json",
				},
			},
		},
		{
			Name:    "Get dashboard - subpath - success",
			TestDir: dir,
			Commands: []Command{
				{
					Command:      "config use-context subpath",
					ExpectedCode: 0,
				},
				{
					Command:        "get Dashboard.ReciqtgGk",
					ExpectedCode:   0,
					ExpectedOutput: "ReciqtgGk.json",
				},
				// Reset context
				{
					Command:      "config use-context default",
					ExpectedCode: 0,
				},
			},
		},
		{
			Name:    "Get dashboard - failure",
			TestDir: dir,
			Commands: []Command{
				{
					Command:      "get missing-dashboard",
					ExpectedCode: 1,
				},
			},
		},
	}

	RunTests(t, tests)
}
