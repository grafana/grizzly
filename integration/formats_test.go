package integration_test

import (
	"fmt"
	"testing"

	"github.com/grafana/grizzly/pkg/testutil"
)

func TestFormats(t *testing.T) {
	ticker := testutil.PingService(testutil.GetUrl())
	defer ticker.Stop()

	tests := []struct {
		Name     string
		PushCmd  string
		Expected string
	}{
		{
			Name:     "Read json/flat/onlyspec, write json/dirs/spec",
			PushCmd:  "push -s -k Dashboard -f general json/flat/onlyspec",
			Expected: "json/dirs/spec",
		},
	}
	grizzlyTests := []GrizzlyTest{}
	for _, test := range tests {
		grizzlyTest := GrizzlyTest{
			Name:    test.Name,
			TestDir: "testdata/formats",
			Commands: []Command{
				{
					Command: "./start-grafana.sh",
				},
				{
					Arguments:     test.PushCmd,
					ExpectedCode:  0,
					ExpectedError: nil,
				},
				{
					Arguments:     "pull temp",
					ExpectedCode:  0,
					ExpectedError: nil,
				},
				{
					Command:       "diff",
					Arguments:     fmt.Sprintf("-ur temp %s", test.Expected),
					ExpectedCode:  0,
					ExpectedError: nil,
				},
			},
		}
		grizzlyTests = append(grizzlyTests, grizzlyTest)
	}
	RunTests(t, grizzlyTests)
}
