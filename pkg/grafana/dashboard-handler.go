package grafana

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"errors"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
	log "github.com/sirupsen/logrus"

	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// Moved from utils.go
const generalFolderId = 0
const generalFolderUID = "general"

// DashboardHandler is a Grizzly Handler for Grafana dashboards
type DashboardHandler struct {
	grizzly.BaseHandler
}

// NewDashboardHandler returns configuration defining a new Grafana Dashboard Handler
func NewDashboardHandler(provider grizzly.Provider) *DashboardHandler {
	return &DashboardHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "Dashboard", true),
	}
}

const (
	dashboardPattern = "dashboards/%s/dashboard-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *DashboardHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(dashboardPattern, resource.GetMetadata("folder"), resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *DashboardHandler) Parse(m map[string]any) (grizzly.Resources, error) {
	resource, err := grizzly.ResourceFromMap(m)
	if err != nil {
		return nil, err
	}
	resource.SetSpecString("uid", resource.Name())
	if !resource.HasMetadata("folder") {
		resource.SetMetadata("folder", generalFolderUID)
	}
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DashboardHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("id")
	resource.DeleteSpecKey("version")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DashboardHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	uid, _ := resource.GetSpecString("uid")
	if uid == "" {
		resource.SetSpecString("uid", resource.Name())
	}
	return &resource
}

// Validate returns the uid of resource
func (h *DashboardHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if resource.Name() != uid && exist {
		return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return nil
}

func (h *DashboardHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	spec := resource["spec"].(map[string]interface{})
	if val, ok := spec["uid"]; ok {
		return val.(string), nil
	}
	return "", fmt.Errorf("UID not specified")
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DashboardHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteDashboard(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %w", UID, err)
	}
	return resource, nil
}

// GetRemote retrieves a dashboard as a resource
func (h *DashboardHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid, _ := resource.GetSpecString("uid")
	if uid != resource.Name() {
		return nil, fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return h.getRemoteDashboard(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *DashboardHandler) ListRemote() ([]string, error) {
	return h.getRemoteDashboardList()
}

// Add pushes a new dashboard to Grafana via the API
func (h *DashboardHandler) Add(resource grizzly.Resource) error {
	resource = *h.Unprepare(resource)
	return h.postDashboard(resource)
}

// Update pushes a dashboard to Grafana via the API
func (h *DashboardHandler) Update(existing, resource grizzly.Resource) error {
	resource = *h.Unprepare(resource)
	return h.postDashboard(resource)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *DashboardHandler) Preview(resource grizzly.Resource, opts *grizzly.PreviewOpts) error {
	s, err := h.postSnapshot(resource, opts)
	if err != nil {
		return err
	}
	notifier.Info(resource, "view: "+s.URL)
	if opts.ExpiresSeconds > 0 {
		notifier.Warn(resource, fmt.Sprintf("Previews will expire and be deleted automatically in %d seconds\n", opts.ExpiresSeconds))
	} else {
		notifier.Error(resource, "delete: "+s.DeleteURL)
	}
	return nil
}

// getRemoteDashboard retrieves a dashboard object from Grafana
func (h *DashboardHandler) getRemoteDashboard(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	dashboardOk, err := client.Dashboards.GetDashboardByUID(uid)
	if err != nil {
		var gErr *dashboards.GetDashboardByUIDNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}
	dashboard := dashboardOk.GetPayload()

	// TODO: Turn spec into a real models.DashboardFullWithMeta object
	spec, err := structToMap(dashboard.Dashboard)
	if err != nil {
		return nil, err
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	folderUid := extractFolderUID(client, *dashboard)
	resource.SetMetadata("folder", folderUid)
	return &resource, nil
}

func (h *DashboardHandler) getRemoteDashboardList() ([]string, error) {
	var (
		limit            = int64(1000)
		searchType       = "dash-db"
		page       int64 = 0
		uids       []string
	)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	params := search.NewSearchParams().WithLimit(&limit).WithType(&searchType)
	for {
		page++
		params.SetPage(&page)

		searchOk, err := client.Search.Search(params, nil)
		if err != nil {
			return nil, err
		}

		for _, hit := range searchOk.GetPayload() {
			uids = append(uids, hit.UID)
		}
		if int64(len(searchOk.GetPayload())) < *params.Limit {
			return uids, nil
		}
	}
}

func (h *DashboardHandler) postDashboard(resource grizzly.Resource) error {
	folderUID := resource.GetMetadata("folder")
	var folderID int64
	if !(folderUID == "General" || folderUID == "general") {
		folderHandler := NewFolderHandler(h.Provider)
		folder, err := folderHandler.getRemoteFolder(folderUID)
		if err != nil {
			if errors.Is(err, grizzly.ErrNotFound) {
				return fmt.Errorf("cannot upload dashboard %s as folder %s not found", resource.Name(), folderUID)
			} else {
				return fmt.Errorf("cannot upload dashboard %s: %w", resource.Name(), err)
			}
		}
		folderID = int64(folder.GetSpecValue("id").(float64))
	} else {
		folderID = generalFolderId
	}

	body := models.SaveDashboardCommand{
		Dashboard: resource.Spec(),
		FolderID:  folderID,
		Overwrite: true,
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	_, err = client.Dashboards.PostDashboard(&body)
	return err
}

func (h *DashboardHandler) postSnapshot(resource grizzly.Resource, opts *grizzly.PreviewOpts) (*models.CreateDashboardSnapshotOKBody, error) {
	body := models.CreateDashboardSnapshotCommand{
		Dashboard: resource.Spec(),
	}
	if opts.ExpiresSeconds > 0 {
		body.Expires = int64(opts.ExpiresSeconds)
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	response, err := client.Snapshots.CreateDashboardSnapshot(&body, nil)
	if err != nil {
		return nil, err
	}
	return response.GetPayload(), nil
}

func (h *DashboardHandler) Detect(data map[string]any) bool {
	expectedKeys := []string{
		"panels",
		"title",
		"schemaVersion",
	}
	for _, key := range expectedKeys {
		_, ok := data[key]
		if !ok {
			return false
		}
	}
	return true
}

func (h *DashboardHandler) GetProxyEndpoints(p grizzly.GrizzlyServer) []grizzly.ProxyEndpoint {
	return []grizzly.ProxyEndpoint{
		{
			Method:  "GET",
			Url:     "/d/{uid}/{slug}",
			Handler: h.RootDashboardPageHandler(p),
		},
		{
			Method:  "GET",
			Url:     "/api/dashboards/uid/{uid}",
			Handler: h.DashboardJSONGetHandler(p),
		},
		{
			Method:  "POST",
			Url:     "/api/dashboards/db/",
			Handler: h.DashboardJSONPostHandler(p),
		},
	}
}

func (h *DashboardHandler) RootDashboardPageHandler(p grizzly.GrizzlyServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		if p.Url == "" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "<p><b>Error:</b> No URL provided")
			return
		}
		req, err := http.NewRequest("GET", p.Url+r.URL.Path, nil)
		if err != nil {
			log.Print(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		req.Header.Set("Authorization", "Bearer "+p.Token)
		req.Header.Set("User-Agent", p.UserAgent)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			w.Write(body)
			return
		}

		msg := ""
		if p.Token == "" {
			msg += "<p><b>Warning:</b> No service account token specified.</p>"
		}

		if resp.StatusCode == 302 {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, msg+"<p>Authentication error</p>")
		} else {
			body, _ := io.ReadAll(resp.Body)
			w.WriteHeader(resp.StatusCode)
			fmt.Fprintf(w, msg+string(body))
		}
	}
}

func (h *DashboardHandler) DashboardJSONGetHandler(p grizzly.GrizzlyServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		if uid == "" {
			http.Error(w, "No UID specified", 400)
			return
		}

		resources, err := p.Parser.Parse()
		if err != nil {
			log.Error("Error: ", err)
			http.Error(w, fmt.Sprintf("Error: %s", err), 500)
			return
		}
		for _, resource := range resources {
			if resource.Kind() == "Dashboard" && resource.Name() == uid {
				meta := map[string]interface{}{
					"type":      "db",
					"isStarred": false,
					"folderID":  0,
					"folderUID": "",
					"url":       fmt.Sprintf("/d/%s/slug", uid),
				}
				wrapper := map[string]interface{}{
					"dashboard": resource.Spec(),
					"meta":      meta,
				}

				out, _ := json.Marshal(wrapper)
				w.Write(out)
				return
			}
		}
		http.Error(w, fmt.Sprintf("Dashboard with UID %s not found", uid), 404)
	}
}

func (h *DashboardHandler) DashboardJSONPostHandler(p grizzly.GrizzlyServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		dash := map[string]interface{}{}
		content, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(content, &dash)
		if err != nil {
			http.Error(w, "Error parsing JSON", 400)
			return
		}

		resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), "dummy", dash)
		if err != nil {
			http.Error(w, "Error creating resource", 400)
			return
		}
		uid, err := h.GetUID(resource)
		if err != nil {
			http.Error(w, "Error getting dashboard UID", 400)
			return
		}
		resource.SetMetadata("name", uid)

		out, _, _, err := grizzly.Format(p.Registry, p.ResourcePath, &resource, p.Opts.OutputFormat, p.Opts.OnlySpec)
		if err != nil {
			http.Error(w, "Error formatting content", 400)
			return
		}

		err = os.WriteFile(p.ResourcePath, out, 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error writing file: %s", err), 400)
			return
		}

		jout := map[string]interface{}{
			"id":      1,
			"slug":    "slug",
			"status":  "success",
			"uid":     uid,
			"url":     fmt.Sprintf("/d/%s/slug", uid),
			"version": 1,
		}
		body, _ := json.Marshal(jout)
		w.Write(body)
	}
}

func (h *DashboardHandler) ProxyURL(resource grizzly.Resource) (string, error) {
	uid, err := h.GetUID(resource)
	return fmt.Sprintf("/d/%s/slug", uid), err
}
