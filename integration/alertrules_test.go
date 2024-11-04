package integration_test

import (
	"testing"
)

func TestAlertRuleGroups(t *testing.T) {
	dir := "testdata/alert-rules"
	setupContexts(t, dir)

	t.Run("Apply rule group", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:      "apply sample-rule-group.json",
					ExpectedCode: 0,
				},
			},
		})
	})

	t.Run("Get previously applied rule group", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					// get [handler].[folder_uid].[rule_group_uid]
					Command:      "get AlertRuleGroup.adxrm7wi8un0gf.test_eval_group",
					ExpectedCode: 0,
				},
			},
		})
	})
}
