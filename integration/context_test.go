package integration_test

import (
	"testing"
)

func TestContexts(t *testing.T) {
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
		},
	}

	RunTests(t, tests)
}
