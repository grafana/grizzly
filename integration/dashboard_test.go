package integration_test

import (
	"testing"

	"github.com/grafana/grizzly/pkg/testutil"
)

func TestDashboard(t *testing.T) {

	ticker := testutil.PingService(testutil.GetUrl())
	defer ticker.Stop()

	tests := []GrizzlyTest{
		{
			Name:    "Get dashboard - success",
			TestDir: "testdata/dashboards",
			Commands: []Command{
				{
					Command:        "get Dashboard.ReciqtgGk",
					ExpectedCode:   0,
					ExpectedError:  nil,
					ExpectedOutput: "ReciqtgGk.json",
				},
			},
			Validate: func(t *testing.T) {

			},
		},
		{
			Name:    "Get dashboard - failure",
			TestDir: "testdata/dashboards",
			Commands: []Command{
				{
					Command:      "get missing-dashboard",
					ExpectedCode: 1,
				},
			},
			Validate: func(t *testing.T) {

			},
		},
	}

	RunTests(t, tests)
}
