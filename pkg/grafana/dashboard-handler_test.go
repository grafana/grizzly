package grafana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/nsf/jsondiff"
)

func TestParseDashboard(t *testing.T) {
	//handler := DashboardHandler{}
	//func Parse(path string, i interface{}) (grizzly.ResourceList, error) {
}

func TestDashboardDiff(t *testing.T) {
	//(h *DashboardHandler) Diff(notifier grizzly.Notifier, resources grizzly.ResourceList) error {

}

func TestDashboardApply(t *testing.T) {
	//func (h *DashboardHandler) Apply(notifier grizzly.Notifier, resources grizzly.ResourceList) error {
}

func TestDashboardGetRemoteRepresentation(t *testing.T) {
	tests := []struct {
		uid      string
		text     string
		resource string
		err      error
	}{
		{
			uid: "prod-overview",
			text: fmt.Sprintf(`{
				"dashboard": %s,
				"meta": {
					"folderId": 0,
					"folderTitle": "General"
				}
			}`, GetTestDashboard("prod-overview")),
			resource: GetTestDashboard("prod-overview"),
			err:      nil,
		},
		{
			uid:  "dashboard-error",
			text: `{"status": "error"}`,
			err:  fmt.Errorf("Error retrieving dashboard"),
		},
	}

	provider := Provider{}
	handler := NewDashboardHandler(provider)
	for _, test := range tests {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if test.err != nil {
				w.WriteHeader(500)
				w.Write([]byte(test.err.Error()))
			} else {
				w.WriteHeader(200)
				w.Write([]byte(test.text))
			}
		}))
		defer srv.Close()
		os.Setenv("GRAFANA_URL", srv.URL)

		resource, err := handler.GetRemoteRepresentation(test.uid)
		if test.err == nil && err != nil {
			t.Errorf("Retrieving dashboard %s returned unexpected error: %v", test.uid, err)
		} else if test.err != nil && err == nil {
			t.Errorf("Retrieving dashboard %s should return error but didn't: %v", test.uid, test.err)
		} else {
			opts := jsondiff.DefaultConsoleOptions()
			diff, msg := jsondiff.Compare([]byte(resource), []byte(test.resource), &opts)
			if diff != jsondiff.FullMatch {
				t.Error(msg)
			}
		}
	}
}

func TestDashboardAdd(t *testing.T) {
	//func (h *DashboardHandler) Add(resource grizzly.Resource) error {
}

func TestDashboardUpdate(t *testing.T) {
	//func (h *DashboardHandler) Update(existing, resource grizzly.Resource) error {
}

func TestDashboardPreview(t *testing.T) {
	//func (h *DashboardHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
}

func TestDashboardListen(t *testing.T) {
	//func (h *DashboardHandler) Listen(notifier grizzly.Notifier, UID, filename string) error {
}

func GetTestDashboard(uid string) string {
	dashboards := map[string]string{
		"prod-overview": `{
		"uid": "prod-overview",
		"title": "Production Overview",
		"tags": ["templated"],
		"timezone": "browser",
		"schemaVersion": 17
	}`,
	}
	return dashboards[uid]
}
