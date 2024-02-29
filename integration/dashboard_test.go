package integration_test

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/grafana/grizzly/pkg/grafana"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestDashboard(t *testing.T) {
	dir := "testdata/dashboards"
	setupContexts(t, dir)

	t.Run("Get dashboard - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:            "get Dashboard.ReciqtgGk",
					ExpectedCode:       0,
					ExpectedOutputFile: "ReciqtgGk.yml",
				},
			},
		})
	})

	t.Run("Get dashboard list - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:            "list -r -t Dashboard",
					ExpectedOutputFile: "list.txt",
				},
			},
		})
	})

	t.Run("Apply dashboard - no folder", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:        "apply no-folder.yml",
					ExpectedOutput: "Dashboard.no-folder added\n",
				},
				{
					Command:                "get Dashboard.no-folder",
					ExpectedOutputContains: "folder: general",
				},
			},
		})
	})

	t.Run("Re-apply dashboard - no folder", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:        "apply no-folder.yml",
					ExpectedOutput: "Dashboard.no-folder no differences\n",
				},
				{
					Command:                "get Dashboard.no-folder",
					ExpectedOutputContains: "folder: general",
				},
			},
		})
	})

	t.Run("Re-apply changed dashboard - no folder", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:        "apply no-folder-mk2.yml",
					ExpectedOutput: "Dashboard.no-folder updated\n",
				},
				{
					Command:                "get Dashboard.no-folder",
					ExpectedOutputContains: "folder: general",
				},
			},
		})
	})

	t.Run("Diff dashboard - success", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:        "diff ReciqtgGk.yml",
					ExpectedCode:   0,
					ExpectedOutput: "Dashboard.ReciqtgGk no differences\n",
				},
			},
		})
	})

	t.Run("Diff dashboard - invalid auth", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: []string{"invalid_auth"},
			Commands: []Command{
				{
					Command:             "diff ReciqtgGk.yml",
					ExpectedCode:        1,
					ExpectedLogsContain: "Invalid username or password",
				},
			},
		})
	})

	t.Run("Get dashboard - failure", func(t *testing.T) {
		runTest(t, GrizzlyTest{
			TestDir:       dir,
			RunOnContexts: allContexts,
			Commands: []Command{
				{
					Command:      "get missing-dashboard",
					ExpectedCode: 1,
				},
			},
		})
	})
}

func TestDashboardHandler(t *testing.T) {
	provider, err := grafana.NewProvider(&testutil.TestContext().Grafana)
	require.NoError(t, err)
	handler := grafana.NewDashboardHandler(provider)

	t.Run("get remote dashboard - success", func(t *testing.T) {
		resource, err := handler.GetByUID("ReciqtgGk")
		require.NoError(t, err)

		require.Equal(t, "grizzly.grafana.com/v1alpha1", resource.APIVersion())
		require.Equal(t, "ReciqtgGk", resource.Name())
		require.NotEmpty(t, resource.GetMetadata("folder"))
	})

	t.Run("get remote dashboard - not found", func(t *testing.T) {
		_, err := handler.GetByUID("dummy")
		require.ErrorContains(t, err, "not found")
	})

	t.Run("get remote dashboard list - success", func(t *testing.T) {
		list, err := handler.ListRemote()
		require.NoError(t, err)

		require.Len(t, list, 4)
		require.EqualValues(t, []string{"ReciqtgGk", "392Ik4GGk", "kE0IIVGGz", "no-folder"}, list)
	})

	t.Run("post remote dashboard - success", func(t *testing.T) {
		wd, _ := os.Getwd()
		log.Printf("PWD: %s", wd)
		dashboard, err := os.ReadFile("testdata/test_json/post_dashboard.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(dashboard, &resource)
		require.NoError(t, err)

		err = handler.Add(resource)
		require.NoError(t, err)

		dash, err := handler.GetByUID("d4sHb0ard-")
		require.NoError(t, err)
		require.NotNil(t, dash)

		require.Equal(t, "abcdefghi", resource.GetMetadata("folder"))
	})

	t.Run("post remote dashboard - not found", func(t *testing.T) {
		resource := map[string]interface{}{
			"metadata": map[string]interface{}{
				"folder": "dummy",
				"name":   "dummy",
			},
			"spec": map[string]interface{}{},
		}

		err := handler.Add(resource)
		require.EqualError(t, err, "cannot upload dashboard dummy as folder dummy not found")
	})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			"metadata": map[string]interface{}{
				"name": "test",
			},
		}
		handler := grafana.DashboardHandler{}
		uid, err := handler.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, "test", uid)
	})

	_ = os.Unsetenv("GRAFANA_URL")

	t.Run("Validate/Prepare", func(t *testing.T) {
		newResource := func(name string, spec map[string]any) grizzly.Resource {
			resource := grizzly.Resource{
				"apiVersion": "apiVersion",
				"kind":       "Dashboard",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			}
			return resource
		}
		handler := grafana.DashboardHandler{}
		tests := []struct {
			Name                string
			Resource            grizzly.Resource
			ValidateErrorString string
			ExpectedUID         string
		}{
			{
				Name:                "name and UID match",
				Resource:            newResource("name1", map[string]any{"uid": "name1"}),
				ValidateErrorString: "",
				ExpectedUID:         "name1",
			},
			{
				Name:                "no UID provided",
				Resource:            newResource("name1", map[string]any{"title": "something"}),
				ValidateErrorString: "",
				ExpectedUID:         "name1",
			},
			{
				Name:                "name and UID differ",
				Resource:            newResource("name1", map[string]any{"uid": "name2"}),
				ValidateErrorString: "uid 'name2' and name 'name1', don't match",
			},
		}
		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				err := handler.Validate(test.Resource)
				if test.ValidateErrorString != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), test.ValidateErrorString)
					return
				}
				require.NoError(t, err)
				newResource := handler.Prepare(nil, test.Resource)
				uid, _ := handler.GetUID(*newResource)
				require.Equal(t, uid, test.ExpectedUID)
			})
		}
	})
}
