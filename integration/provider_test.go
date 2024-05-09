package integration_test

import (
	"testing"
)

func TestEmptyProviders(t *testing.T) {
	dir := "testdata/providers"
	setupContexts(t, dir)

	t.Run("Providers work for offline commands", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: dir,
			Commands: []Command{
				{
					Command:      "config use-context empty",
					ExpectedCode: 0,
				},
				{
					Command:            "show folder.yaml",
					ExpectedCode:       0,
					ExpectedOutputFile: "folder-output.txt",
				},
			},
		})
	})

	t.Run("Providers fail online commands without configs", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir: dir,
			Commands: []Command{
				{
					Command: "config use-context empty",
				},
				{
					Command:      "apply folder.yaml",
					ExpectedCode: 1,
				},
			},
		})
	})

}
