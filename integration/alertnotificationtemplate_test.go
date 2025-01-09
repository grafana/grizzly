package integration_test

import (
	"testing"
)

func TestAlertNotificationTemplates(t *testing.T) {
	dir := "testdata/alertnotificationtemplates"
	setupContexts(t, dir)

	t.Run("Get template - not found", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:      "get AlertNotificationTemplate.dummy",
					ExpectedCode: 1,
				},
			},
		})
	})

	t.Run("Apply template", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:      "apply standard-template.yaml",
					ExpectedCode: 0,
				},
			},
		})
	})

	t.Run("Apply same template", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:                "apply standard-template.yaml",
					ExpectedCode:           0,
					ExpectedOutputContains: "AlertNotificationTemplate.standard-template unchanged\n",
				},
			},
		})
	})

	t.Run("Get applied template", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:            "get AlertNotificationTemplate.standard-template",
					ExpectedCode:       0,
					ExpectedOutputFile: "standard-template.yaml",
				},
			},
		})
	})

	t.Run("Get remote templates list", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:            "list -r -t AlertNotificationTemplate",
					ExpectedCode:       0,
					ExpectedOutputFile: "list.txt",
				},
			},
		})
	})

	t.Run("Diff template with no differences", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:                "diff standard-template.yaml",
					ExpectedCode:           0,
					ExpectedOutputContains: "AlertNotificationTemplate.standard-template no differences",
				},
			},
		})
	})
}
