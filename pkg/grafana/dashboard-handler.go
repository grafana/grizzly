package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/grizzly/notifier"
)

// Moved from utils.go
const generalFolderID = 0
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

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *DashboardHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("id")
	resource.DeleteSpecKey("version")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *DashboardHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	if !resource.HasSpecString("uid") {
		resource.SetSpecString("uid", resource.Name())
	}
	if !resource.HasMetadata("folder") {
		resource.SetMetadata("folder", generalFolderUID)
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
	uid, ok := resource.GetSpecString("uid")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}
	return uid, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *DashboardHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteDashboard(uid)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %w", uid, err)
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

// Snapshot pushes dashboards as snapshots
func (h *DashboardHandler) Snapshot(resource grizzly.Resource, expiresSeconds int) error {
	s, err := h.postSnapshot(resource, expiresSeconds)
	if err != nil {
		return err
	}
	notifier.Info(resource, "view: "+s.URL)
	if expiresSeconds > 0 {
		notifier.Warn(resource, fmt.Sprintf("Snapshots will expire and be deleted automatically in %d seconds\n", expiresSeconds))
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
	folderUID := extractFolderUID(client, *dashboard)
	resource.SetMetadata("folder", folderUID)
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
	if !(folderUID == DefaultFolder || folderUID == strings.ToLower(DefaultFolder)) {
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
		folderID = generalFolderID
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

func (h *DashboardHandler) postSnapshot(resource grizzly.Resource, expiresSeconds int) (*models.CreateDashboardSnapshotOKBody, error) {
	body := models.CreateDashboardSnapshotCommand{
		Dashboard: &models.Unstructured{Object: resource.Spec()},
	}
	if expiresSeconds > 0 {
		body.Expires = int64(expiresSeconds)
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

func (h *DashboardHandler) GetProxyEndpoints(p grizzly.Server) []grizzly.ProxyEndpoint {
	return []grizzly.ProxyEndpoint{
		{
			Method:  "GET",
			URL:     "/d/{uid}/{slug}",
			Handler: h.resourceFromQueryParameterMiddleware(p, "grizzly_from_file", h.RootDashboardPageHandler(p)),
		},
		{
			Method:  "GET",
			URL:     "/api/dashboards/uid/{uid}",
			Handler: h.DashboardJSONGetHandler(p),
		},
		{
			Method:  "POST",
			URL:     "/api/dashboards/db",
			Handler: h.DashboardJSONPostHandler(p),
		},
		{
			Method:  "POST",
			URL:     "/api/dashboards/db/",
			Handler: h.DashboardJSONPostHandler(p),
		},
	}
}

func (h *DashboardHandler) resourceFromQueryParameterMiddleware(p grizzly.Server, parameterName string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fromFilePath := r.URL.Query().Get(parameterName); fromFilePath != "" {
			if err := p.ParseResources(fromFilePath); err != nil {
				grizzly.SendError(w, "could not parse resource", fmt.Errorf("could not parse resource"), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

func (h *DashboardHandler) RootDashboardPageHandler(p grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		config := h.Provider.(ClientProvider).Config()
		if config.URL == "" {
			grizzly.SendError(w, "Error: No Grafana URL configured", fmt.Errorf("no Grafana URL configured"), 400)
			return
		}
		req, err := http.NewRequest("GET", config.URL+r.URL.Path, nil)
		if err != nil {
			grizzly.SendError(w, http.StatusText(500), err, 500)
			return
		}
		req.Header.Set("Authorization", "Bearer "+config.Token)
		req.Header.Set("User-Agent", p.UserAgent)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			writeOrLog(w, body)
			return
		}

		msg := ""
		if config.Token == "" {
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

func (h *DashboardHandler) DashboardJSONGetHandler(p grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		if uid == "" {
			grizzly.SendError(w, "No UID specified", fmt.Errorf("no UID specified within the URL"), 400)
			return
		}

		resource, found := p.Resources.Find(grizzly.NewResourceRef("Dashboard", uid))
		if !found {
			grizzly.SendError(w, fmt.Sprintf("Dashboard with UID %s not found", uid), fmt.Errorf("dashboard with UID %s not found", uid), 404)
			return
		}
		if resource.GetSpecValue("version") == nil {
			resource.SetSpecValue("version", 1)
		}
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
		writeOrLog(w, out)
	}
}

func (h *DashboardHandler) DashboardJSONPostHandler(p grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Dashboard map[string]any `json:"dashboard"`
		}{}
		content, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(content, &resp)
		if err != nil {
			grizzly.SendError(w, "Error parsing JSON", err, 400)
			return
		}
		resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), "dummy", resp.Dashboard)
		if err != nil {
			grizzly.SendError(w, "Error creating resource", err, 400)
			return
		}
		uid, err := h.GetSpecUID(resource)
		if err != nil {
			grizzly.SendError(w, "Error getting dashboard UID", err, 400)
			return
		}
		if uid == "" {
			grizzly.SendError(w, "Dashboard has no UID", fmt.Errorf("dashboard has no UID"), 400)
			return
		}
		resource.SetMetadata("name", uid)
		resource.SetSpecString("uid", uid)

		out, _, _, err := grizzly.Format(p.Registry, p.ResourcePath, &resource, p.OutputFormat, p.OnlySpec)
		if err != nil {
			grizzly.SendError(w, "Error formatting content", err, 500)
			return
		}

		existing, found := p.Resources.Find(grizzly.NewResourceRef("Dashboard", uid))
		if !found {
			grizzly.SendError(w, fmt.Sprintf("Dashboard with UID %s not found", uid), fmt.Errorf("dashboard with UID %s not found", uid), 500)
			return
		}
		if !existing.Source.Rewritable {
			grizzly.SendError(w, "The source for this dashboard is not rewritable", fmt.Errorf("the source for this dashboard is not rewritable"), 400)
			return
		}

		err = os.WriteFile(existing.Source.Path, out, 0644)
		if err != nil {
			grizzly.SendError(w, fmt.Sprintf("Error writing file: %s", err), err, 500)
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
		writeOrLog(w, body)
	}
}

func (h *DashboardHandler) ProxyURL(resource grizzly.Resource) (string, error) {
	uid, err := h.GetUID(resource)
	return fmt.Sprintf("/d/%s/slug", uid), err
}
