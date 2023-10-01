package grafana

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	gerrors "github.com/go-openapi/errors"
	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

// Losing a bunch of omitempty fields

// getRemoteDatasource retrieves a datasource object from Grafana
func getRemoteDatasource(client *gclient.GrafanaHTTPAPI, uid string) (*grizzly.Resource, error) {
	h := DatasourceHandler{}

	params := datasources.NewGetDataSourceByUIDParams().WithUID(uid)
	datasourceOk, err := client.Datasources.GetDataSourceByUID(params, nil)
	var datasource *models.DataSource
	if err != nil {
		var gErr gerrors.Error
		if errors.As(err, &gErr) && gErr.Code() == http.StatusNotFound {
			params := datasources.NewGetDataSourceByNameParams().WithName(uid)
			datasourceOk, err := client.Datasources.GetDataSourceByName(params, nil)
			if err != nil {
				if errors.As(err, &gErr) && gErr.Code() == http.StatusNotFound {
					return nil, grizzly.ErrNotFound
				}
				return nil, err
			} else {
				datasource = datasourceOk.GetPayload()
			}
		} else {
			return nil, err
		}
	} else {
		datasource = datasourceOk.GetPayload()
	}

	// TODO: Turn spec into a real models.Datasource object
	spec, err := structToMap(datasource)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	return &resource, nil
}

func getRemoteDatasourceList(client *gclient.GrafanaHTTPAPI) ([]string, error) {
	params := datasources.NewGetDataSourcesParams()
	datasourcesOk, err := client.Datasources.GetDataSources(params, nil)
	if err != nil {
		return nil, err
	}
	datasources := datasourcesOk.GetPayload()

	uids := make([]string, len(datasources))
	for i, datasource := range datasources {
		uids[i] = datasource.UID
	}
	return uids, nil
}

func postDatasource(client *gclient.GrafanaHTTPAPI, resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.DataSource object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var datasource models.AddDataSourceCommand
	err = json.Unmarshal(data, &datasource)
	if err != nil {
		return err
	}
	params := datasources.NewAddDataSourceParams().WithBody(&datasource)
	_, err = client.Datasources.AddDataSource(params, nil)
	return err
}

func putDatasource(client *gclient.GrafanaHTTPAPI, resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.DataSource object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var modelDatasource models.DataSource
	err = json.Unmarshal(data, &modelDatasource)
	if err != nil {
		return err
	}

	var datasource models.UpdateDataSourceCommand
	err = json.Unmarshal(data, &datasource)
	if err != nil {
		return err
	}
	params := datasources.NewUpdateDataSourceByIDParams().WithID(strconv.FormatInt(modelDatasource.ID, 10)).WithBody(&datasource)
	_, err = client.Datasources.UpdateDataSourceByID(params, nil)
	return err
}
