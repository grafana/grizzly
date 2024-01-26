package main

import (
	"testing"

	"github.com/grafana/grizzly/pkg/testutil"
)

func TestContexts(t *testing.T) {

	ticker := testutil.PingService(testutil.GetUrl())
	defer ticker.Stop()

	tests := []GrizzlyTest{
		{
			Name:    "Get contexts - success",
			TestDir: "testdata/contexts",
			Commands: []Command{
				{
					Command:        "config get-contexts",
					ExpectedCode:   0,
					ExpectedError:  nil,
					ExpectedOutput: "get-contexts.txt",
				},
			},
			Validate: func(t *testing.T) {

			},
		},
	}

	RunTests(t, tests)
}
