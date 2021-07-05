package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestDashboard(t *testing.T) {
	if os.Getenv("CI") != "" {
		return
	}
	os.Setenv("GRAFANA_URL", "http://localhost:3000")

	go func() {
		cmd := exec.Command("/bin/sh", "run_container.sh")
		if o, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("fail to spin up container, err: %q", o)
			os.Exit(1)
		}
	}()

	defer func() {
		cmd := exec.Command("/bin/sh", "kill_container.sh")
		if o, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("fail to kill container, err: %q", o)
			os.Exit(1)
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutExceeded := time.After(120 * time.Second)

	success := false
	for !success {
		select {
		case <-timeoutExceeded:
			fmt.Println("failed")

		case <-ticker.C:
			resp, _ := http.Get("http://localhost:3000/")
			if resp != nil {
				success = true
				break
			}
		}
	}

	t.Run("get remote dashboard - success", func(t *testing.T) {
		resource, err := getRemoteDashboard("ReciqtgGk")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "ReciqtgGk")
		require.NotEmpty(t, resource.GetMetadata("folder"))
	})

	t.Run("get remote dashboard - not found", func(t *testing.T) {
		_, err := getRemoteDashboard("dummy")
		require.EqualError(t, err, "not found")
	})

	t.Run("get remote dashboard list - success", func(t *testing.T) {
		list, err := getRemoteDashboardList()
		require.NoError(t, err)

		require.Len(t, list, 3)
		require.EqualValues(t, []string{"ReciqtgGk", "392Ik4GGk", "kE0IIVGGz"}, list)
	})

	t.Run("post remote dashboard - success", func(t *testing.T) {
		dashboard, err := os.ReadFile("testdata/test_json/post_dashboard.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(dashboard, &resource)
		require.NoError(t, err)

		err = postDashboard(resource)
		require.NoError(t, err)

		dash, err := getRemoteDashboard("d4sHb0ard-")
		require.NoError(t, err)
		require.NotNil(t, dash)

		require.Equal(t, resource.GetMetadata("folder"), "abcdefghi")
	})

	t.Run("post remote dashboard - not found", func(t *testing.T) {
		var resource grizzly.Resource
		resource = map[string]interface{}{
			"metadata": map[string]interface{}{
				"folder": "dummy",
			},
		}

		err := postDashboard(resource)
		require.EqualError(t, err, "couldn't fetch folder 'dummy' from remote: not found")
	})

	_ = os.Unsetenv("GRAFANA_URL")
}
