package integration_test

import (
	"testing"
)

func TestContexts(t *testing.T) {
	setupContexts(t, "testdata/contexts")

	t.Run("Get contexts - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: "testdata/contexts",
			Commands: []Command{
				{
					Command:            "config get-contexts",
					ExpectedCode:       0,
					ExpectedOutputFile: "get-contexts.txt",
				},
			},
		})
	})
}
