package grafana

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var _ grizzly.ProxyConfigurator = &datasourceProxyConfigurator{}

// datasourceProxyConfigurator describes how to proxy Datasource resources.
type datasourceProxyConfigurator struct {
	provider grizzly.Provider
}

func (c *datasourceProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/connections/datasources/edit/%s", uid)
}

func (c *datasourceProxyConfigurator) ProxyEditURL(uid string) string {
	return c.ProxyURL(uid)
}

func (c *datasourceProxyConfigurator) Endpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/connections/datasources/edit/{uid}",
			Handler: authenticateAndProxyHandler(s, c.provider),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/datasources/uid/{uid}",
			Handler: c.datasourceJSONGetHandler(s),
		},
	}
}

func (c *datasourceProxyConfigurator) StaticEndpoints() grizzly.StaticProxyConfig {
	return grizzly.StaticProxyConfig{
		ProxyGet: []string{
			"/api/instance/plugins",
			"/api/instance/provisioned-plugins",
			"/api/plugins",
			"/api/plugin-proxy/*",
			"/api/usage/datasource/*",
		},
	}
}

func (c *datasourceProxyConfigurator) datasourceJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		if uid == "" {
			httputils.Error(w, "No UID specified", fmt.Errorf("no UID specified within the URL"), http.StatusBadRequest)
			return
		}

		resource, found := s.Resources.Find(grizzly.NewResourceRef(DatasourceKind, uid))
		if !found {
			httputils.Error(w, fmt.Sprintf("Datasource with UID %s not found", uid), fmt.Errorf("datasource with UID %s not found", uid), http.StatusNotFound)
			return
		}

		// These values are required for the page to load properly.
		if resource.GetSpecValue("version") == nil {
			resource.SetSpecValue("version", 1)
		}
		if resource.GetSpecValue("id") == nil {
			resource.SetSpecValue("id", 1)
		}

		// we don't support saving datasources via the proxy yet
		resource.SetSpecValue("readOnly", true)

		// to remove some "missing permissions warning" and enable some features
		resource.SetSpecValue("accessControl", map[string]any{
			"datasources.caching:read":      true,
			"datasources.caching:write":     false,
			"datasources.id:read":           true,
			"datasources.permissions:read":  true,
			"datasources.permissions:write": true,
			"datasources:delete":            false,
			"datasources:query":             true,
			"datasources:read":              true,
			"datasources:write":             true,
		})

		httputils.WriteJSON(w, resource.Spec())
	}
}
