package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

const DatasourceKind = "Datasource"

var _ grizzly.Handler = &DatasourceHandler{}
var _ grizzly.ProxyConfiguratorProvider = &DatasourceHandler{}

// DatasourceHandler is a Grizzly Handler for Grafana datasources
type DatasourceHandler struct {
	grizzly.BaseHandler
}

// NewDatasourceHandler returns a new Grizzly Handler for Grafana datasources
func NewDatasourceHandler(provider grizzly.Provider) *DatasourceHandler {
	return &DatasourceHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, DatasourceKind, false),
	}
}

const (
	datasourcePattern = "datasources/datasource-%s.%s"
)

// ProxyConfigurator provides a configurator object describing how to proxy datasources.
func (h *DatasourceHandler) ProxyConfigurator() grizzly.ProxyConfigurator {
	return &datasourceProxyConfigurator{
		provider: h.Provider,
	}
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *DatasourceHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	filename := strings.ReplaceAll(resource.Name(), string(os.PathSeparator), "-")
	return fmt.Sprintf(datasourcePattern, filename, filetype)
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DatasourceHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("version")
	resource.DeleteSpecKey("id")
	resource.DeleteSpecKey("secureJsonData")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DatasourceHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	if existing != nil && existing.GetSpecValue("id") != nil {
		resource.SetSpecValue("id", existing.GetSpecValue("id"))
	}
	resource.DeleteSpecKey("version")
	if !resource.HasSpecString("uid") {
		resource.SetSpecValue("uid", resource.Name())
	}
	return &resource
}

// Validate returns the uid of resource
func (h *DatasourceHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return ErrUIDNameMismatch{UID: uid, Name: resource.Name()}
		}
	}
	return nil
}

func (h *DatasourceHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	uid, ok := resource.GetSpecString("uid")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}

	return uid, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DatasourceHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return h.getRemoteDatasource(uid)
}

// GetRemote retrieves a datasource as a Resource
func (h *DatasourceHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteDatasource(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *DatasourceHandler) ListRemote() ([]string, error) {
	return h.getRemoteDatasourceList()
}

// Add pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Add(resource grizzly.Resource) error {
	return h.postDatasource(resource)
}

// Update pushes a datasource to Grafana via the API
func (h *DatasourceHandler) Update(existing, resource grizzly.Resource) error {
	return h.putDatasource(resource)
}

// getRemoteDatasource retrieves a datasource object from Grafana
func (h *DatasourceHandler) getRemoteDatasource(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	datasourceOk, err := client.Datasources.GetDataSourceByUID(uid)
	var datasource *models.DataSource
	if err != nil {
		var gErr *datasources.GetDataSourceByUIDNotFound
		if errors.As(err, &gErr) {
			datasourceOk, err := client.Datasources.GetDataSourceByName(uid, nil)
			if err != nil {
				// OpenAPI definition does not define 404 for GetDataSourceByName, so falls though to runtime.APIError.
				var gErr *runtime.APIError
				if errors.As(err, &gErr) && gErr.IsCode(http.StatusNotFound) {
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

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (h *DatasourceHandler) getRemoteDatasourceList() ([]string, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	response, err := client.Datasources.GetDataSources()
	if err != nil {
		return nil, err
	}

	uids := make([]string, len(response.GetPayload()))
	for i, datasource := range response.GetPayload() {
		uids[i] = datasource.UID
	}
	return uids, nil
}

func (h *DatasourceHandler) postDatasource(resource grizzly.Resource) error {
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

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	_, err = client.Datasources.AddDataSource(&datasource, nil)
	return err
}

func (h *DatasourceHandler) putDatasource(resource grizzly.Resource) error {
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

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	_, err = client.Datasources.UpdateDataSourceByID(strconv.FormatInt(modelDatasource.ID, 10), &datasource)
	return err
}
