package grafana

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestSortFolders(t *testing.T) {
	handler := NewFolderHandler(&Provider{})
	folder := func(uid string, parentUID string) grizzly.Resource {
		spec := map[string]interface{}{
			"uid": uid,
		}
		if parentUID != "" {
			spec["parentUid"] = parentUID
		}
		resource, _ := grizzly.NewResource(handler.APIVersion(), handler.Kind(), uid, spec)
		return resource
	}

	cases := []struct {
		name     string
		folders  []grizzly.Resource
		expected []string // expected order of UIDs
	}{
		{
			name:     "empty",
			folders:  []grizzly.Resource{},
			expected: []string{},
		},
		{
			name: "no parents",
			folders: []grizzly.Resource{
				folder("a", ""),
				folder("b", ""),
				folder("c", ""),
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "one parent",
			folders: []grizzly.Resource{
				folder("b", "a"),
				folder("a", ""),
				folder("c", ""),
			},
			expected: []string{"a", "c", "b"},
		},
		{
			name: "nested",
			folders: []grizzly.Resource{
				folder("c", "b"),
				folder("b", "a"),
				folder("a", ""),
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "nested with siblings",
			folders: []grizzly.Resource{
				folder("d", "b"),
				folder("c", "b"),
				folder("b", "a"),
				folder("a", ""),
			},
			expected: []string{"a", "b", "d", "c"},
		},
		{
			name: "parent not declared",
			folders: []grizzly.Resource{
				folder("a", "c"),
				folder("b", ""),
			},
			expected: []string{"a", "b"},
		},
		{
			name: "parent not declared - nested",
			folders: []grizzly.Resource{
				folder("b", "a"),
				folder("a", "c"),
			},
			expected: []string{"a", "b"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sorted := handler.Sort(grizzly.NewResources(tc.folders...))
			require.Equal(t, len(tc.expected), sorted.Len())
			for i, resource := range sorted.AsList() {
				require.Equal(t, tc.expected[i], resource.Name())
			}
		})
	}
}
