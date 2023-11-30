package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/grafana/grizzly/pkg/testutil"
	"github.com/stretchr/testify/require"
)

type GrizzlyTest struct {
	Name         string
	ConfigFile   string
	Commands     []string
	ExpectedCode int
	Validate     func(t *testing.T)
}

func TestDashboard(t *testing.T) {

	ticker := testutil.PingService(testutil.GetUrl())
	defer ticker.Stop()

	tests := []GrizzlyTest{
		{
			Name:       "Get dashboard - success",
			ConfigFile: "test1.yaml",
			Commands: []string{
				"get ReciqtgGk",
			},
			ExpectedCode: 0,
			Validate: func(t *testing.T) {

			},
		},
		{
			Name:       "Get dashboard - failure",
			ConfigFile: "test1.yaml",
			Commands: []string{
				"get missing-dashboard",
			},
			ExpectedCode: -1,
			Validate: func(t *testing.T) {

			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			for _, commandstr := range test.Commands {
				args := []string{}
				args = append(args, "--context", test.ConfigFile)
				args = append(args, strings.Split(commandstr, " ")...)
				cmd := exec.Command("./grr", args...)
				fmt.Println(cmd.CombinedOutput())
				exitCode := cmd.ProcessState.ExitCode()
				require.Equal(t, test.ExpectedCode, exitCode, "Exited with %d (%d expected)", exitCode, test.ExpectedCode)
			}
			test.Validate(t)
		})
	}
}
