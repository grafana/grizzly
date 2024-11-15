package grafana

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var _ grizzly.ProxyConfigurator = &dashboardProxyConfigurator{}

// dashboardProxyConfigurator describes how to proxy Dashboard resources.
type dashboardProxyConfigurator struct {
	provider grizzly.Provider
}

func (c *dashboardProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/d/%s/slug", uid)
}

func (c *dashboardProxyConfigurator) GetProxyEndpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/d/{uid}/{slug}",
			Handler: c.resourceFromQueryParameterMiddleware(s, "grizzly_from_file", authenticateAndProxyHandler(s, c.provider)),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/dashboards/uid/{uid}",
			Handler: c.dashboardJSONGetHandler(s),
		},
		{
			Method:  http.MethodPost,
			URL:     "/api/dashboards/db",
			Handler: c.dashboardJSONPostHandler(s),
		},
		{
			Method:  http.MethodPost,
			URL:     "/api/dashboards/db/",
			Handler: c.dashboardJSONPostHandler(s),
		},
	}
}

func (c *dashboardProxyConfigurator) resourceFromQueryParameterMiddleware(s grizzly.Server, parameterName string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fromFilePath := r.URL.Query().Get(parameterName); fromFilePath != "" {
			if _, err := s.ParseResources(fromFilePath); err != nil {
				httputils.Error(w, "could not parse resource", fmt.Errorf("could not parse resource"), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

func (c *dashboardProxyConfigurator) dashboardJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		if uid == "" {
			httputils.Error(w, "No UID specified", fmt.Errorf("no UID specified within the URL"), http.StatusBadRequest)
			return
		}

		resource, found := s.Resources.Find(grizzly.NewResourceRef("Dashboard", uid))
		if !found {
			httputils.Error(w, fmt.Sprintf("Dashboard with UID %s not found", uid), fmt.Errorf("dashboard with UID %s not found", uid), http.StatusNotFound)
			return
		}
		if resource.GetSpecValue("version") == nil {
			resource.SetSpecValue("version", 1)
		}

		httputils.WriteJSON(w, map[string]any{
			"dashboard": resource.Spec(),
			"meta": map[string]any{
				"type":      "db",
				"isStarred": false,
				"folderID":  0,
				"folderUID": "",
				"url":       c.ProxyURL(uid),
			},
		})
	}
}

func (c *dashboardProxyConfigurator) dashboardJSONPostHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Dashboard map[string]any `json:"dashboard"`
		}{}
		content, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(content, &resp); err != nil {
			httputils.Error(w, "Error parsing JSON", err, http.StatusBadRequest)
			return
		}
		uid, ok := resp.Dashboard["uid"].(string)
		if !ok || uid == "" {
			httputils.Error(w, "Dashboard has no UID", fmt.Errorf("dashboard has no UID"), http.StatusBadRequest)
			return
		}
		resource, ok := s.Resources.Find(grizzly.NewResourceRef(DashboardKind, uid))
		if !ok {
			err := fmt.Errorf("unknown dashboard: %s", uid)
			httputils.Error(w, err.Error(), err, http.StatusBadRequest)
			return
		}

		resource.SetSpec(resp.Dashboard)

		if err := s.UpdateResource(uid, resource); err != nil {
			httputils.Error(w, err.Error(), err, http.StatusInternalServerError)
			return
		}

		httputils.WriteJSON(w, map[string]any{
			"id":      1,
			"slug":    "slug",
			"status":  "success",
			"uid":     uid,
			"url":     fmt.Sprintf("/d/%s/slug", uid),
			"version": 1,
		})
	}
}
