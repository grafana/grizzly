package grizzly_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	. "github.com/grafana/grizzly/pkg/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPull(t *testing.T) {
	provider := grafana.NewProviderWithConfig(GetTestConfig())
	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			provider,
		})

	ticker := PingService(GetUrl())
	defer ticker.Stop()

	opts := grizzly.Opts{
		Targets: []string{
			"Datasource/392IktgGk",
		},
	}

	t.Run("with existing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), filepath.Base(t.Name()))
		f, err := os.Create(path)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		err = grizzly.Pull(path, opts)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "pull <resource-path> must be a directory")
	})

	t.Run("with existing folder", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), filepath.Base(t.Name()))
		err := os.MkdirAll(path, 0755)
		require.NoError(t, err)

		err = grizzly.Pull(path, opts)
		assert.NoError(t, err)
		assert.Equal(t, 1, numOfFiles(path))
	})

	t.Run("with non-existing folder", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), filepath.Base(t.Name()))
		err := grizzly.Pull(path, opts)
		assert.NoError(t, err)
		assert.Equal(t, 1, numOfFiles(path))
	})
}

func numOfFiles(path string) int {
	files, _ := os.ReadDir(path)
	return len(files)
}
