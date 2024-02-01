package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Command struct {
	Command        string
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
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			for _, command := range test.Commands {
				args := []string{}
				args = append(args, strings.Split(command.Command, " ")...)
				cwd, _ := os.Getwd()
				stdErrBuf := &strings.Builder{}
				cmd := exec.Command(filepath.Join(cwd, "../grr"), args...)
				cmd.Stderr = stdErrBuf
				cmd.Dir = test.TestDir
				output, err := cmd.Output()
				if command.ExpectedError != nil {
					require.Error(t, err, command.ExpectedError)
				}
				exitCode := cmd.ProcessState.ExitCode()
				require.Equal(t, command.ExpectedCode, exitCode, "Exited with %d (%d expected).\nOutput: %s\nstderr: %s", exitCode, command.ExpectedCode, string(output), stdErrBuf.String())
				if command.ExpectedOutput != "" {
					data, err := os.ReadFile(filepath.Join(test.TestDir, command.ExpectedOutput))
					require.NoError(t, err)
					require.Equal(t, string(data), string(output))
				}
			}
			if test.Validate != nil {
				test.Validate(t)
			}
		})
	}
}

// Setting up a settings.yaml file. We don't want to commit this file to avoid git diff.
func setupContexts(t *testing.T, dir string) {
	_, err := os.Create(filepath.Join(dir, "settings.yaml"))
	require.NoError(t, err)
	RunTests(t, []GrizzlyTest{
		{
			Name:    "Create contexts",
			TestDir: dir,
			Commands: []Command{
				{Command: "config create-context subpath"},
				{Command: "config use-context subpath"},
				{Command: "config set grafana.url http://localhost:3003/grafana"},
				{Command: "config create-context default"},
				{Command: "config use-context default"},
				{Command: "config set grafana.url http://localhost:3001"},
			},
		},
	})

	t.Cleanup(func() {
		assert.NoError(t, os.Remove(filepath.Join(dir, "settings.yaml")))
	})
}
