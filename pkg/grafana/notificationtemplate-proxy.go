package grafana

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var _ grizzly.ProxyConfigurator = &alertNotificationTemplateProxyConfigurator{}

// alertNotificationTemplateProxyConfigurator describes how to proxy AlertNotificationTemplate resources.
type alertNotificationTemplateProxyConfigurator struct {
	provider grizzly.Provider
}

func (c *alertNotificationTemplateProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/alerting/notifications/templates/%s/edit", uid)
}

func (c *alertNotificationTemplateProxyConfigurator) Endpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/alerting/notifications/templates/{template_uid}/edit",
			Handler: authenticateAndProxyHandler(s, c.provider),
		},
		// Depending on the Grafana version, the frontend can call either of these endpoints
		{
			Method:  http.MethodGet,
			URL:     "/apis/notifications.alerting.grafana.app/v0alpha1/namespaces/{namespace}/templategroups/{template_uid}",
			Handler: c.templateGetAsK8S(s),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/alertmanager/grafana/config/api/v1/alerts",
			Handler: c.alertManagerConfigGet(s),
		},
	}
}

func (c *alertNotificationTemplateProxyConfigurator) StaticEndpoints() grizzly.StaticProxyConfig {
	return grizzly.StaticProxyConfig{
		ProxyPost: []string{
			"/api/alertmanager/grafana/config/api/v1/templates/test",
		},
	}
}

func (c *alertNotificationTemplateProxyConfigurator) alertManagerConfigGet(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates := s.Resources.OfKind(KindAlertNotificationTemplate).AsList()

		templatesMap := make(map[string]any, len(templates))
		for _, template := range templates {
			templatesMap[template.Name()] = template.GetSpecValue("template")
		}

		httputils.WriteJSON(w, map[string]any{
			"template_files":      templatesMap,
			"alertmanager_config": map[string]any{},
		})
	}
}

func (c *alertNotificationTemplateProxyConfigurator) templateGetAsK8S(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateUID := chi.URLParam(r, "template_uid")

		template, found := s.Resources.Find(grizzly.NewResourceRef(KindAlertNotificationTemplate, templateUID))
		if !found {
			httputils.Error(w, fmt.Sprintf("Alert notification template with UID %s not found", templateUID), fmt.Errorf("alert notification template with UID %s not found", templateUID), http.StatusNotFound)
			return
		}

		httputils.WriteJSON(w, map[string]any{
			"kind":       "TemplateGroup",
			"apiVersion": "notifications.alerting.grafana.app/v0alpha1",
			"metadata": map[string]any{
				"name":            templateUID,
				"uid":             templateUID,
				"namespace":       chi.URLParam(r, "namespace"),
				"resourceVersion": "resource-version",
			},
			"spec": map[string]any{
				"title":   templateUID,
				"content": template.GetSpecValue("template"),
			},
		})
	}
}
