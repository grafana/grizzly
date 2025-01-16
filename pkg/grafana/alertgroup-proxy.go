package grafana

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var _ grizzly.ProxyConfigurator = &alertRuleProxyConfigurator{}

// alertRuleProxyConfigurator describes how to proxy AlertRuleGroup resources.
type alertRuleProxyConfigurator struct {
	provider grizzly.Provider
}

func (c *alertRuleProxyConfigurator) ProxyURL(uid string) string {
	return fmt.Sprintf("/alerting/grafana/%s/view", uid)
}

func (c *alertRuleProxyConfigurator) ProxyEditURL(uid string) string {
	return fmt.Sprintf("/alerting/%s/edit", uid)
}

func (c *alertRuleProxyConfigurator) Endpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/alerting/grafana/{rule_uid}/view",
			Handler: authenticateAndProxyHandler(s, c.provider),
		},
		{
			Method:  http.MethodGet,
			URL:     "/alerting/grafana/{rule_uid}/edit",
			Handler: authenticateAndProxyHandler(s, c.provider),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/ruler/grafana/api/v1/rule/{rule_uid}",
			Handler: c.alertRuleJSONGetHandler(s),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/ruler/grafana/api/v1/rules/{folder_uid}/{rule_group_uid}",
			Handler: c.alertRuleGroupJSONGetHandler(s),
		},
	}
}

func (c *alertRuleProxyConfigurator) StaticEndpoints() grizzly.StaticProxyConfig {
	return grizzly.StaticProxyConfig{
		ProxyGet: []string{
			"/api/v1/ngalert",
		},
		ProxyPost: []string{
			"/api/v1/eval",
		},
	}
}

func (c *alertRuleProxyConfigurator) alertRuleGroupJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folderUID := chi.URLParam(r, "folder_uid")
		ruleGroupUID := chi.URLParam(r, "rule_group_uid")
		fullUID := joinAlertRuleGroupUID(folderUID, ruleGroupUID)

		ruleGroup, found := s.Resources.Find(grizzly.NewResourceRef(AlertRuleGroupKind, fullUID))
		if !found {
			httputils.Error(w, fmt.Sprintf("Alert rule group with UID %s not found", fullUID), fmt.Errorf("alert rule group with UID %s not found", fullUID), http.StatusNotFound)
			return
		}

		interval := time.Duration(ruleGroup.GetSpecValue("interval").(float64)) * time.Second

		rules := ruleGroup.GetSpecValue("rules").([]any)
		formattedRules := make([]map[string]any, 0, len(rules))
		for _, rule := range rules {
			formattedRules = append(formattedRules, toGrafanaAlert(rule.(map[string]any), interval))
		}

		httputils.WriteJSON(w, map[string]any{
			"name":     ruleGroup.GetSpecValue("title"),
			"interval": interval.String(),
			"rules":    formattedRules,
		})
	}
}

func (c *alertRuleProxyConfigurator) alertRuleJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ruleUID := chi.URLParam(r, "rule_uid")
		if ruleUID == "" {
			httputils.Error(w, "No alert rule UID specified", fmt.Errorf("no alert rule UID specified within the URL"), http.StatusBadRequest)
			return
		}

		var rule map[string]any
		var ruleGroup grizzly.Resource
		ruleFound := false
		_ = s.Resources.OfKind(AlertRuleGroupKind).ForEach(func(candidate grizzly.Resource) error {
			if ruleFound {
				return nil
			}

			rules := candidate.GetSpecValue("rules").([]any)
			for _, candidateRule := range rules {
				candidateUID := candidateRule.(map[string]any)["uid"].(string)
				if candidateUID != ruleUID {
					continue
				}

				ruleFound = true
				rule = candidateRule.(map[string]any)
				ruleGroup = candidate
			}

			return nil
		})
		if !ruleFound {
			httputils.Error(w, fmt.Sprintf("Alert rule with UID %s not found", ruleUID), fmt.Errorf("rule group with UID %s not found", ruleUID), http.StatusNotFound)
			return
		}
		interval := time.Duration(ruleGroup.GetSpecValue("interval").(float64)) * time.Second

		httputils.WriteJSON(w, toGrafanaAlert(rule, interval))
	}
}

// See GettableGrafanaRule model in grafana-openapi-client-go
func toGrafanaAlert(rule map[string]any, ruleGroupInterval time.Duration) map[string]any {
	var version any = 1
	if v, ok := rule["version"]; ok {
		version = v
	}

	intervalSeconds := 0
	interval, err := time.ParseDuration(rule["for"].(string))
	if err == nil {
		intervalSeconds = int(interval.Seconds())
	}

	grafanaAlert := rule
	grafanaAlert["intervalSeconds"] = intervalSeconds
	grafanaAlert["version"] = version
	grafanaAlert["namespace_uid"] = rule["folderUID"]
	grafanaAlert["rule_group"] = rule["ruleGroup"]
	grafanaAlert["no_data_state"] = rule["noDataState"]
	grafanaAlert["exec_err_state"] = rule["execErrState"]
	grafanaAlert["is_paused"] = false
	grafanaAlert["metadata"] = map[string]any{
		"editor_settings": map[string]any{},
	}

	return map[string]any{
		"expr":          "",
		"for":           ruleGroupInterval.String(),
		"grafana_alert": grafanaAlert,
	}
}
