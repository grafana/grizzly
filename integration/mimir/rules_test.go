package mimir

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/mimir"
	"github.com/grafana/grizzly/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRules(t *testing.T) {
	dirName := "testdata/rules"
	provider := mimir.NewProvider(&testutil.TestContext().Mimir)
	handler := provider.GetHandlers()[0]

	t.Run("create rule group", func(t *testing.T) {
		dirs, err := os.ReadDir(dirName)
		require.NoError(t, err)

		for _, dir := range dirs {
			file, err := os.ReadFile(filepath.Join(dirName, dir.Name()))
			require.NoError(t, err)

			var resource grizzly.Resource
			require.NoError(t, yaml.Unmarshal(file, &resource.Body))
			assert.NoError(t, handler.Add(resource))
		}
	})

	// Mimir takes some seconds in sync the rules. If we get the list of them immediately, it could return an empty list.
	time.Sleep(1500 * time.Millisecond)

	t.Run("get remote rule list", func(t *testing.T) {
		ids, err := handler.ListRemote()
		require.NoError(t, err)
		fixedIDs := make([]string, len(ids))
		for i, id := range ids {
			fixedIDs[i] = strings.Split(id, ".")[1]
		}

		sort.Slice(fixedIDs, func(i, j int) bool {
			return fixedIDs[i] < fixedIDs[j]
		})

		assert.Equal(t, fixedIDs, []string{"test-rules-1", "test-rules-2", "test-rules-3"})
	})
}
