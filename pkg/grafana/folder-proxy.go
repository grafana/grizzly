package grafana

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var _ grizzly.ProxyConfigurator = &folderProxyConfigurator{}

// folderProxyConfigurator describes how to proxy DashboardFolder resources.
type folderProxyConfigurator struct {
	provider grizzly.Provider
}

func (c *folderProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/dashboards/f/%s/", uid)
}

func (c *folderProxyConfigurator) Endpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/alerting/{rule_uid}/edit",
			Handler: authenticateAndProxyHandler(s, c.provider),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/folders/{folder_uid}",
			Handler: c.folderJSONGetHandler(s),
		},
	}
}

func (c *folderProxyConfigurator) StaticEndpoints() grizzly.StaticProxyConfig {
	return grizzly.StaticProxyConfig{}
}

func (c *folderProxyConfigurator) folderJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folderUID := chi.URLParam(r, "folder_uid")
		withAccessControl := r.URL.Query().Get("accesscontrol")

		folder, found := s.Resources.Find(grizzly.NewResourceRef(DashboardFolderKind, folderUID))
		if !found {
			httputils.Error(w, fmt.Sprintf("Folder with UID %s not found", folderUID), fmt.Errorf("folder with UID %s not found", folderUID), http.StatusNotFound)
			return
		}

		// These values are required for the page to load properly.
		if folder.GetSpecValue("version") == nil {
			folder.SetSpecValue("version", 1)
		}
		if folder.GetSpecValue("id") == nil {
			folder.SetSpecValue("id", 1)
		}

		response := folder.Spec()

		if withAccessControl == "true" {
			// TODO: can we omit stuff from this list?
			response["accessControl"] = map[string]any{
				"alert.rules:create":           false,
				"alert.rules:delete":           false,
				"alert.rules:read":             true,
				"alert.rules:write":            false,
				"alert.silences:create":        false,
				"alert.silences:read":          true,
				"alert.silences:write":         false,
				"annotations:create":           false,
				"annotations:delete":           false,
				"annotations:read":             true,
				"annotations:write":            false,
				"dashboards.permissions:read":  true,
				"dashboards.permissions:write": false,
				"dashboards:create":            true,
				"dashboards:delete":            false,
				"dashboards:read":              true,
				"dashboards:write":             true,
				"folders.permissions:read":     true,
				"folders.permissions:write":    false,
				"folders:create":               false,
				"folders:delete":               false,
				"folders:read":                 true,
				"folders:write":                false,
				"library.panels:create":        false,
				"library.panels:delete":        false,
				"library.panels:read":          true,
				"library.panels:write":         false,
			}
		}

		httputils.WriteJSON(w, response)
	}
}
