package grafana

import (
	"encoding/json"
	"strings"

	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Losing typeLogoUrl, version, withCredentials, secureJsonFields
// Losing a bunch of omitempty fields

// getRemoteDatasource retrieves a datasource object from Grafana
func getRemoteDatasource(uid string) (*grizzly.Resource, error) {
	h := DatasourceHandler{}
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	datasource, err := client.DataSourceByUID(uid)
	// TODO: Restore lookup by name functionality, underlying library lacks it
	if err != nil {
		if strings.HasPrefix(err.Error(), "status: 404") {
			return nil, grizzly.ErrNotFound
		}
	}

	// TODO: Turn spec into a real gapi.Datasource object
	var spec map[string]interface{}
	data, err := json.Marshal(datasource)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	return &resource, nil
}

func getRemoteDatasourceList() ([]string, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	datasources, err := client.DataSources()
	if err != nil {
		return nil, err
	}

	uids := make([]string, len(datasources))
	for i, datasource := range datasources {
		uids[i] = datasource.UID
	}
	return uids, nil
}

func postDatasource(resource grizzly.Resource) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// TODO: Turn spec into a real gapi.DataSource object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var datasource gapi.DataSource
	err = json.Unmarshal(data, &datasource)
	if err != nil {
		return err
	}
	_, err = client.NewDataSource(&datasource)
	return err
}

func putDatasource(resource grizzly.Resource) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// TODO: Turn spec into a real gapi.DataSource object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var datasource gapi.DataSource
	err = json.Unmarshal(data, &datasource)
	if err != nil {
		return err
	}
	return client.UpdateDataSource(&datasource)
}
