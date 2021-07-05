package grafana

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestDatasources(t *testing.T) {
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

	t.Run("get remote datasource - success", func(t *testing.T) {
		resource, err := getRemoteDatasource("AppDynamics")
		require.NoError(t, err)

		require.Equal(t, resource.APIVersion(), "grizzly.grafana.com/v1alpha1")
		require.Equal(t, resource.Name(), "AppDynamics")
		require.Len(t, resource.Spec(), 20)
	})

	t.Run("get remote datasource - not found", func(t *testing.T) {
		_, err := getRemoteDatasource("dummy")
		require.EqualError(t, err, "not found")
	})

	t.Run("get remote datasources list", func(t *testing.T) {
		resources, err := getRemoteDatasourceList()
		require.NoError(t, err)

		require.NotNil(t, resources)
		require.Len(t, resources, 1)
	})

	t.Run("post remote datasource - success", func(t *testing.T) {
		datasource, err := os.ReadFile("testdata/test_json/post_datasource.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(datasource, &resource)
		require.NoError(t, err)

		err = postDatasource(resource)
		require.NoError(t, err)

		ds, err := getRemoteDatasource("appdynamics")
		require.NoError(t, err)
		require.NotNil(t, ds)
		require.Equal(t, ds.Spec()["type"], "dlopes7-appdynamics-datasource")

		t.Run("put remote datasource - update", func(t *testing.T) {
			ds.SetSpecString("type", "new-type")

			err := putDatasource(*ds)
			require.NoError(t, err)

			updatedDS, err := getRemoteDatasource("appdynamics")
			require.NoError(t, err)

			require.Equal(t, updatedDS.Spec()["type"], "new-type")
		})
	})

	t.Run("post remote datasource - conflict - datasource already exists", func(t *testing.T) {
		datasource, err := os.ReadFile("testdata/test_json/post_datasource.json")
		require.NoError(t, err)

		var resource grizzly.Resource

		err = json.Unmarshal(datasource, &resource)
		require.NoError(t, err)

		resource.SetSpecString("name", "AppDynamics")

		err = postDatasource(resource)

		grafanaErr := err.(ErrNon200Response)
		require.Error(t, err)
		require.Equal(t, grafanaErr.Response.StatusCode, 409)
	})
}
