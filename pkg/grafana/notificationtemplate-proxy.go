package grafana

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/internal/utils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var _ grizzly.ProxyConfigurator = &alertNotificationTemplateProxyConfigurator{}

// alertNotificationTemplateProxyConfigurator describes how to proxy AlertNotificationTemplate resources.
type alertNotificationTemplateProxyConfigurator struct {
	provider grizzly.Provider
}

// ProxyURL returns the URL to use to view an alerting notification templates via the proxy.
func (c *alertNotificationTemplateProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/alerting/notifications/templates/%s/edit", uid)
}

// Endpoints lists HTTP handlers to register on the proxy.
func (c *alertNotificationTemplateProxyConfigurator) Endpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/alerting/notifications/templates/{template_uid}/edit",
			Handler: authenticateAndProxyHandler(s, c.provider),
		},
		// Depending on the Grafana version, the frontend might call k8s-style endpoints
		{
			Method:  http.MethodGet,
			URL:     "/apis/notifications.alerting.grafana.app/v0alpha1/namespaces/{namespace}/templategroups",
			Handler: c.listAsK8S(s),
		},
		{
			Method:  http.MethodGet,
			URL:     "/apis/notifications.alerting.grafana.app/v0alpha1/namespaces/{namespace}/templategroups/{template_uid}",
			Handler: c.getAsK8S(s),
		},
		{
			Method:  http.MethodPut,
			URL:     "/apis/notifications.alerting.grafana.app/v0alpha1/namespaces/{namespace}/templategroups/{template_uid}",
			Handler: c.saveAsK8S(s),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/alertmanager/grafana/config/api/v1/alerts",
			Handler: c.alertManagerConfigGet(s),
		},
		{
			Method:  http.MethodPost,
			URL:     "/api/alertmanager/grafana/config/api/v1/alerts",
			Handler: c.alertManagerConfigSave(s),
		},
	}
}

// StaticEndpoints lists endpoints to be proxied transparently.
func (c *alertNotificationTemplateProxyConfigurator) StaticEndpoints() grizzly.StaticProxyConfig {
	return grizzly.StaticProxyConfig{
		ProxyGet: []string{
			"/api/alertmanager/grafana/api/v2/alerts",
		},
		ProxyPost: []string{
			"/api/alertmanager/grafana/config/api/v1/templates/test",
		},
	}
}

// alertManagerConfigGet serves a partially mocked alert manager config to the UI.
// Only the templates are served from Grizzly resources.
func (c *alertNotificationTemplateProxyConfigurator) alertManagerConfigGet(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates := s.Resources.OfKind(KindAlertNotificationTemplate).AsList()

		templatesMap := make(map[string]any, len(templates))
		for _, template := range templates {
			templatesMap[template.Name()] = template.GetSpecValue("template")
		}

		// The frontend expects most of these values :|
		httputils.WriteJSON(w, map[string]any{
			"template_files": templatesMap,
			"alertmanager_config": map[string]any{
				"route": map[string]any{
					"receiver": "dummy-receiver",
					"group_by": []string{"grafana_folder", "alertname"},
					"routes":   []map[string]any{},
				},
				"receivers": []map[string]any{
					{
						"name": "dummy-receiver",
						"grafana_managed_receiver_configs": []map[string]any{
							{
								"uid":                   "dummy-receiver",
								"name":                  "dummy-receiver",
								"type":                  "email",
								"disableResolveMessage": false,
								"settings": map[string]any{
									"addresses": "\u003cexample@email.com\u003e",
								},
								"secureFields": map[string]any{},
							},
						},
					},
				},
			},
		})
	}
}

// alertManagerConfigSave persists an alert manager config edited via the UI.
// Only the templates are persisted as Grizzly resources.
func (c *alertNotificationTemplateProxyConfigurator) alertManagerConfigSave(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := &struct {
			TemplateFiles map[string]string `json:"template_files"`
		}{}
		content, err := io.ReadAll(r.Body)
		if err != nil {
			httputils.Error(w, "could not read request body", err, http.StatusInternalServerError)
		}
		if err := json.Unmarshal(content, input); err != nil {
			httputils.Error(w, "Error parsing request", err, http.StatusBadRequest)
			return
		}

		for uid, templateContent := range input.TemplateFiles {
			template, found := s.Resources.Find(grizzly.NewResourceRef(KindAlertNotificationTemplate, uid))
			if !found {
				c.httpNotFound(w, uid)
				return
			}

			template.SetSpecValue("template", templateContent)

			if err := s.UpdateResource(template); err != nil {
				httputils.Error(w, err.Error(), err, http.StatusInternalServerError)
				return
			}
		}

		httputils.WriteJSON(w, map[string]string{
			"message": "configuration created",
		})
	}
}

func (c *alertNotificationTemplateProxyConfigurator) listAsK8S(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates := s.Resources.OfKind(KindAlertNotificationTemplate).AsList()
		templatesAsK8S := utils.Map(templates, func(template grizzly.Resource) map[string]any {
			return c.resourceAsK8S(template.Name(), chi.URLParam(r, "namespace"), template)
		})

		httputils.WriteJSON(w, map[string]any{
			"kind":       "TemplateGroupList",
			"apiVersion": "notifications.alerting.grafana.app/v0alpha1",
			"metadata":   map[string]any{},
			"items":      templatesAsK8S,
		})
	}
}

func (c *alertNotificationTemplateProxyConfigurator) getAsK8S(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateUID := chi.URLParam(r, "template_uid")

		template, found := s.Resources.Find(grizzly.NewResourceRef(KindAlertNotificationTemplate, templateUID))
		if !found {
			c.httpNotFound(w, templateUID)
			return
		}

		httputils.WriteJSON(w, c.resourceAsK8S(templateUID, chi.URLParam(r, "namespace"), template))
	}
}

func (c *alertNotificationTemplateProxyConfigurator) saveAsK8S(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateUID := chi.URLParam(r, "template_uid")

		template, ok := s.Resources.Find(grizzly.NewResourceRef(KindAlertNotificationTemplate, templateUID))
		if !ok {
			c.httpNotFound(w, templateUID)
			return
		}

		input := &struct {
			Metadata map[string]any `json:"metadata"`
			Spec     struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			} `json:"spec"`
		}{}
		content, err := io.ReadAll(r.Body)
		if err != nil {
			httputils.Error(w, "could not read request body", err, http.StatusInternalServerError)
		}
		if err := json.Unmarshal(content, input); err != nil {
			httputils.Error(w, "Error parsing request", err, http.StatusBadRequest)
			return
		}

		template.SetSpecValue("template", input.Spec.Content)

		if err := s.UpdateResource(template); err != nil {
			httputils.Error(w, err.Error(), err, http.StatusInternalServerError)
			return
		}

		httputils.WriteJSON(w, c.resourceAsK8S(templateUID, chi.URLParam(r, "namespace"), template))
	}
}

func (c *alertNotificationTemplateProxyConfigurator) httpNotFound(w http.ResponseWriter, uid string) {
	httputils.Error(w, fmt.Sprintf("Alert notification template with UID %s not found", uid), fmt.Errorf("alert notification template with UID %s not found", uid), http.StatusNotFound)
}

func (c *alertNotificationTemplateProxyConfigurator) resourceAsK8S(uid string, namespace string, resource grizzly.Resource) map[string]any {
	return map[string]any{
		"kind":       "TemplateGroup",
		"apiVersion": "notifications.alerting.grafana.app/v0alpha1",
		"metadata": map[string]any{
			"name":            uid,
			"uid":             uid,
			"namespace":       namespace,
			"resourceVersion": "resource-version",
		},
		"spec": map[string]any{
			"title":   uid,
			"content": resource.GetSpecValue("template"),
		},
	}
}
