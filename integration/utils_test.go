package integration_test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Command struct {
	Command        string
	Arguments      string
	ExpectedCode   int
	ExpectedError  error
	ExpectedOutput string
}
type GrizzlyTest struct {
	Name     string
	TestDir  string
	Commands []Command
	Validate func(t *testing.T)
}

func RunTests(t *testing.T, tests []GrizzlyTest) {
	cwd, _ := os.Getwd()
	grrCmd := filepath.Join(cwd, "../grr")
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			for _, command := range test.Commands {
				log.Printf("CMD: %s", command.Arguments)
				args := []string{}
				args = append(args, strings.Split(command.Arguments, " ")...)
				var cmd *exec.Cmd
				if command.Command == "" {
					cmd = exec.Command(grrCmd, args...)
				} else {
					cmd = exec.Command(command.Command, args...)
				}
				cmd.Dir = test.TestDir
				output, err := cmd.CombinedOutput()
				if command.ExpectedError != nil {
					require.Error(t, err, command.ExpectedError)
				}
				if command.ExpectedOutput != "" {
					data, err := os.ReadFile(filepath.Join(test.TestDir, command.ExpectedOutput))
					require.NoError(t, err, fmt.Sprintf(""))
					require.Contains(t, string(output), string(data))
				}
				log.Printf("OUT: %s", output)
				exitCode := cmd.ProcessState.ExitCode()
				require.Equal(t, command.ExpectedCode, exitCode, "Exited with %d (%d expected)", exitCode, command.ExpectedCode)
			}
			if test.Validate != nil {
				test.Validate(t)
			}
		})
	}
}
