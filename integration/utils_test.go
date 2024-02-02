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

var allContexts = []string{"default", "subpath", "basic_auth"}

type Command struct {
	Command                string
	ExpectedCode           int
	ExpectedError          error
	ExpectedLogsContain    string
	ExpectedOutput         string
	ExpectedOutputFile     string
	ExpectedOutputContains string
}
type GrizzlyTest struct {
	TestDir       string
	Commands      []Command
	Validate      func(t *testing.T)
	RunOnContexts []string
}

func runTest(t *testing.T, test GrizzlyTest) {
	t.Helper()

	contexts := []string{"default"}
	if len(test.RunOnContexts) > 0 {
		contexts = test.RunOnContexts
	}
	for _, context := range contexts {
		t.Run(context, func(t *testing.T) {
			commands := []Command{
				{Command: "config use-context " + context},
			}
			commands = append(commands, test.Commands...)
			for _, command := range commands {
				stdout, stderr, err := runLocalGrizzly(t, test.TestDir, command.Command)
				exitCode := 0
				if err != nil {
					exitCode = err.(*exec.ExitError).ExitCode()
				}
				require.Equal(t, command.ExpectedCode, exitCode, "Exited with %d (%d expected).\nOutput: %s\nstderr: %s", exitCode, command.ExpectedCode, stdout, stderr)

				// Check stderr
				if command.ExpectedLogsContain != "" {
					require.Contains(t, stderr, command.ExpectedLogsContain)
				}

				// Check stdout
				if command.ExpectedOutputFile != "" {
					bytes, err := os.ReadFile(filepath.Join(test.TestDir, command.ExpectedOutputFile))
					require.NoError(t, err)
					command.ExpectedOutput = string(bytes)
				}
				if command.ExpectedOutput != "" {
					require.Equal(t, command.ExpectedOutput, stdout)
				}
				if command.ExpectedOutputContains != "" {
					require.Contains(t, stdout, command.ExpectedOutputContains)
				}
			}
			if test.Validate != nil {
				test.Validate(t)
			}
		})
	}
}

func runLocalGrizzly(t *testing.T, dir string, command string) (stdout, stderr string, err error) {
	t.Helper()

	args := []string{}
	args = append(args, strings.Split(command, " ")...)
	cwd, _ := os.Getwd()
	stdErrBuf := &strings.Builder{}
	cmd := exec.Command(filepath.Join(cwd, "../grr"), args...)
	cmd.Stderr = stdErrBuf
	cmd.Dir = dir
	output, err := cmd.Output()
	return string(output), stdErrBuf.String(), err
}

// Setting up a settings.yaml file. We don't want to commit this file to avoid git diff.
func setupContexts(t *testing.T, dir string) {
	t.Helper()

	_, err := os.Create(filepath.Join(dir, "settings.yaml"))
	require.NoError(t, err)
	for _, command := range []string{
		"config create-context default",
		"config use-context default",
		"config set grafana.url http://localhost:3001",

		"config create-context subpath",
		"config use-context subpath",
		"config set grafana.url http://localhost:3003/grafana",

		"config create-context basic_auth",
		"config use-context basic_auth",
		"config set grafana.url http://localhost:3004",
		"config set grafana.user admin",
		"config set grafana.token admin",

		"config create-context invalid_auth",
		"config use-context invalid_auth",
		"config set grafana.url http://localhost:3004",
		"config set grafana.user admin",
		"config set grafana.token invalid",
	} {
		_, _, err = runLocalGrizzly(t, dir, command)
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		assert.NoError(t, os.Remove(filepath.Join(dir, "settings.yaml")))
	})
}
