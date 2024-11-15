package grafana

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
	log "github.com/sirupsen/logrus"
)

var _ grizzly.ProxyConfigurator = &libraryElementProxyConfigurator{}

// libraryElementProxyConfigurator describes how to proxy LibraryElement resources.
type libraryElementProxyConfigurator struct {
	provider grizzly.Provider
}

func (c *libraryElementProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/api/library-elements/%s", uid)
}

func (c *libraryElementProxyConfigurator) GetProxyEndpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  "GET",
			URL:     "/api/library-elements/{uid}",
			Handler: c.libraryElementJSONGetHandler(s),
		},
	}
}

func (c *libraryElementProxyConfigurator) libraryElementJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		if uid == "" {
			httputils.Error(w, "No UID specified", fmt.Errorf("no UID specified within the URL"), http.StatusBadRequest)
			return
		}

		resource, found := s.Resources.Find(grizzly.NewResourceRef(LibraryElementKind, uid))
		if !found {
			log.Debug("Library element not found in memory, proxying request to remote Grafana", "uid", uid)
			s.ProxyRequestHandler(w, r)
			return
		}

		if resource.GetSpecValue("version") == nil {
			resource.SetSpecValue("version", 1)
		}

		httputils.WriteJSON(w, map[string]any{
			"result": resource.Spec(),
		})
	}
}
