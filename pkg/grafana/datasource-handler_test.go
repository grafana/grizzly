package grafana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/nsf/jsondiff"
)

func TestParseDatasource(t *testing.T) {
	//handler := DatasourceHandler{}
	//func Parse(path string, i interface{}) (grizzly.ResourceList, error) {
}

func TestDatasourceDiff(t *testing.T) {
	//(h *DatasourceHandler) Diff(notifier grizzly.Notifier, resources grizzly.ResourceList) error {

}

func TestDatasourceApply(t *testing.T) {
	//func (h *DatasourceHandler) Apply(notifier grizzly.Notifier, resources grizzly.ResourceList) error {
}

func TestDatasourceGetRemoteRepresentation(t *testing.T) {
	tests := []struct {
		uid  string
		text string
		err  error
	}{
		{
			uid:  "prometheus",
			text: GetTestDatasource("prometheus"),
			err:  nil,
		},
		{
			uid:  "prometheus",
			text: `{"status": "error"}`,
			err:  fmt.Errorf("500 Internal Server Error"),
		},
	}

	provider := Provider{}
	handler := NewDatasourceHandler(provider)
	for _, test := range tests {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if test.err != nil {
				w.WriteHeader(500)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(test.text))
			}
		}))
		defer srv.Close()
		os.Setenv("GRAFANA_URL", srv.URL)

		resource, err := handler.GetRemoteRepresentation(test.uid)
		if test.err == nil && err != nil {
			t.Errorf("Retrieving datasource %s returned unexpected error: %v", test.uid, err)
		} else if test.err != nil && err == nil {
			t.Errorf("Retrieving datasource %s should return error but didn't: %v", test.uid, test.err)
		} else if err != nil {
			diff := deep.Equal(err, test.err)
			if diff != nil {
				t.Error(diff)
			}
		} else {
			opts := jsondiff.DefaultConsoleOptions()
			diff, msg := jsondiff.Compare([]byte(resource), []byte(test.text), &opts)
			if diff != jsondiff.FullMatch {
				t.Log(resource)
				t.Error(msg)
			}
		}
	}
}

func TestDatasourceAdd(t *testing.T) {
	//func (h *DatasourceHandler) Add(resource grizzly.Resource) error {
}

func TestDatasourceUpdate(t *testing.T) {
	//func (h *DatasourceHandler) Update(existing, resource grizzly.Resource) error {
}

func TestDatasourcePreview(t *testing.T) {
	//func (h *DatasourceHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
}

func TestDatasourceListen(t *testing.T) {
	//func (h *DatasourceHandler) Listen(notifier grizzly.Notifier, UID, filename string) error {
}

func GetTestDatasource(uid string) string {
	datasources := map[string]string{
		"prometheus": `{
			"access": "proxy",
			"isDefault": true,
			"jsonData": {
			  "httpMethod": "GET"
			},
			"name": "prometheus",
			"type": "prometheus",
			"url": "http://localhost/prometheus/"
		  }`,
	}
	return datasources[uid]
}
