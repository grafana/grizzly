package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

const AlertRuleGroupKind = "AlertRuleGroup"

// AlertRuleGroupHandler is a Grizzly Handler for Grafana alertRuleGroups
type AlertRuleGroupHandler struct {
	grizzly.BaseHandler
}

// NewAlertRuleGroupHandler returns a new Grizzly Handler for Grafana alertRuleGroups
func NewAlertRuleGroupHandler(provider grizzly.Provider) *AlertRuleGroupHandler {
	return &AlertRuleGroupHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, AlertRuleGroupKind, false),
	}
}

const (
	alertRuleGroupPattern = "alert-rules/alertRuleGroup-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertRuleGroupHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	filename := strings.ReplaceAll(resource.Name(), string(os.PathSeparator), "-")
	return fmt.Sprintf(alertRuleGroupPattern, filename, filetype)
}

// Validate checks that the uid format is valid
func (h *AlertRuleGroupHandler) Validate(resource grizzly.Resource) error {
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var group models.AlertRuleGroup
	err = json.Unmarshal(data, &group)
	if err != nil {
		return err
	}
	uid := h.getUID(group)
	if uid != resource.Name() {
		return fmt.Errorf("title/folder combination '%s' and name '%s', don't match", uid, resource.Name())
	}
	return nil
}

func (h *AlertRuleGroupHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	name, ok := resource.GetSpecString("name")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}
	return name, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertRuleGroupHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return h.getRemoteAlertRuleGroup(uid)
}

// GetRemote retrieves a alertRuleGroup as a Resource
func (h *AlertRuleGroupHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteAlertRuleGroup(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertRuleGroupHandler) ListRemote() ([]string, error) {
	return h.getRemoteAlertRuleGroupList()
}

// Add pushes a alertRuleGroup to Grafana via the API
func (h *AlertRuleGroupHandler) Add(resource grizzly.Resource) error {
	return h.createAlertRuleGroup(resource)
}

// Update pushes a alertRuleGroup to Grafana via the API
func (h *AlertRuleGroupHandler) Update(existing, resource grizzly.Resource) error {
	return h.putAlertRuleGroup(existing, resource)
}

func (h *AlertRuleGroupHandler) GetProxyEndpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  http.MethodGet,
			URL:     "/alerting/grafana/{rule_uid}/view",
			Handler: authenticateAndProxyHandler(s, h.Provider),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/ruler/grafana/api/v1/rule/{rule_uid}",
			Handler: h.AlertRuleJSONGetHandler(s),
		},
		{
			Method:  http.MethodGet,
			URL:     "/api/ruler/grafana/api/v1/rules/{folder_uid}/{rule_group_uid}",
			Handler: h.AlertRuleGroupJSONGetHandler(s),
		},
	}
}

func (h *AlertRuleGroupHandler) ProxyURL(uid string) string {
	return fmt.Sprintf("/alerting/grafana/%s/view", uid)
}

func (h *AlertRuleGroupHandler) AlertRuleGroupJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folderUID := chi.URLParam(r, "folder_uid")
		ruleGroupUID := chi.URLParam(r, "rule_group_uid")
		fullUID := h.joinUID(folderUID, ruleGroupUID)

		ruleGroup, found := s.Resources.Find(grizzly.NewResourceRef(AlertRuleGroupKind, fullUID))
		if !found {
			grizzly.SendError(w, fmt.Sprintf("Alert rule group with UID %s not found", fullUID), fmt.Errorf("alert rule group with UID %s not found", fullUID), http.StatusNotFound)
			return
		}

		interval := time.Duration(ruleGroup.GetSpecValue("interval").(int)) * time.Second

		rules := ruleGroup.GetSpecValue("rules").([]any)
		formattedRules := make([]map[string]any, 0, len(rules))
		for _, rule := range rules {
			formattedRules = append(formattedRules, toGrafanaAlert(rule.(map[string]any), interval))
		}

		writeJSONOrLog(w, map[string]any{
			"name":     ruleGroup.GetSpecValue("title"),
			"interval": interval.String(),
			"rules":    formattedRules,
		})
	}
}

func (h *AlertRuleGroupHandler) AlertRuleJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ruleUID := chi.URLParam(r, "rule_uid")
		if ruleUID == "" {
			grizzly.SendError(w, "No alert rule UID specified", fmt.Errorf("no alert rule UID specified within the URL"), http.StatusBadRequest)
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
			grizzly.SendError(w, fmt.Sprintf("Alert rule with UID %s not found", ruleUID), fmt.Errorf("rule group with UID %s not found", ruleUID), http.StatusNotFound)
			return
		}

		interval := time.Duration(ruleGroup.GetSpecValue("interval").(int)) * time.Second

		writeJSONOrLog(w, toGrafanaAlert(rule, interval))
	}
}

// getRemoteAlertRuleGroup retrieves a alertRuleGroup object from Grafana
func (h *AlertRuleGroupHandler) getRemoteAlertRuleGroup(uid string) (*grizzly.Resource, error) {
	folder, group := h.splitUID(uid)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	alertRuleGroupOk, err := client.Provisioning.GetAlertRuleGroup(group, folder)
	if err != nil {
		var gErr *provisioning.GetAlertRuleGroupNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}
	alertRuleGroup := alertRuleGroupOk.GetPayload()
	// TODO: Turn spec into a real models.ProvisionedAlertRuleGroup object
	spec, err := structToMap(alertRuleGroup)
	if err != nil {
		return nil, err
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (h *AlertRuleGroupHandler) getRemoteAlertRuleGroupList() ([]string, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	alertRuleGroupsOk, err := client.Provisioning.GetAlertRules()
	if err != nil {
		return nil, err
	}
	alerts := alertRuleGroupsOk.GetPayload()

	uidmap := make(map[string]struct{})
	for _, alert := range alerts {
		uid := h.joinUID(*alert.FolderUID, *alert.RuleGroup)
		uidmap[uid] = struct{}{}
	}
	uids := make([]string, len(uidmap))
	idx := 0
	for k := range uidmap {
		uids[idx] = k
		idx++
	}
	return uids, nil
}

func (h *AlertRuleGroupHandler) createAlertRule(rule *models.ProvisionedAlertRule) error {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	params := provisioning.NewPostAlertRuleParams().WithBody(rule).WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PostAlertRule(params, nil)
	return err
}

func (h *AlertRuleGroupHandler) createAlertRuleGroup(resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.AlertRuleGroup object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var group models.AlertRuleGroup
	if err := json.Unmarshal(data, &group); err != nil {
		return err
	}

	for _, r := range group.Rules {
		if err := h.createAlertRule(r); err != nil {
			return fmt.Errorf("creating rule for group %s: %w", resource.Name(), err)
		}
	}

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(&group).
		WithGroup(group.Title).
		WithFolderUID(group.FolderUID).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutAlertRuleGroup(params)
	return err
}

func (h *AlertRuleGroupHandler) updateAlertRule(rule *models.ProvisionedAlertRule) error {
	rule.ID = 0 // ensure clear id, these should never be used as they are instance-local

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	if rule.UID != "" {
		_, err = client.Provisioning.GetAlertRule(rule.UID)
		if err != nil {
			var gErr *provisioning.GetAlertRuleNotFound
			if errors.As(err, &gErr) {
				return h.createAlertRule(rule)
			}
			return fmt.Errorf("fetching alert rule: %w", err)
		}
	} else {
		params := provisioning.NewPostAlertRuleParams().
			WithBody(rule).
			WithXDisableProvenance(&stringtrue)
		_, err = client.Provisioning.PostAlertRule(params, nil)
		return err
	}

	params := provisioning.NewPutAlertRuleParams().
		WithUID(rule.UID).
		WithBody(rule).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutAlertRule(params)
	return err
}

func unmarshalAlertRuleGroup(resource grizzly.Resource) (*models.AlertRuleGroup, error) {
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return nil, err
	}
	var group models.AlertRuleGroup
	err = json.Unmarshal(data, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func fillAlertRuleGroupUIDs(existing, resource grizzly.Resource) (*models.AlertRuleGroup, error) {
	existingGroup, err := unmarshalAlertRuleGroup(existing)
	if err != nil {
		return nil, err
	}
	t := make(map[string]string)
	for _, rule := range existingGroup.Rules {
		t[*rule.Title] = rule.UID
	}

	updatedGroup, err := unmarshalAlertRuleGroup(resource)
	if err != nil {
		return nil, err
	}

	for _, rule := range updatedGroup.Rules {
		if uid, ok := t[*rule.Title]; ok {
			rule.UID = uid
		}
	}
	return updatedGroup, nil
}

func (h *AlertRuleGroupHandler) putAlertRuleGroup(existing, resource grizzly.Resource) error {
	group, err := fillAlertRuleGroupUIDs(existing, resource)
	if err != nil {
		return err
	}
	for _, r := range group.Rules {
		if err := h.updateAlertRule(r); err != nil {
			return err
		}
	}

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(group).
		WithGroup(group.Title).
		WithFolderUID(group.FolderUID).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutAlertRuleGroup(params, nil)
	return err
}

func (h *AlertRuleGroupHandler) getUID(group models.AlertRuleGroup) string {
	return h.joinUID(group.FolderUID, group.Title)
}

func (h *AlertRuleGroupHandler) joinUID(folder, title string) string {
	return fmt.Sprintf("%s.%s", folder, title)
}

func (h *AlertRuleGroupHandler) splitUID(uid string) (string, string) {
	spl := strings.SplitN(uid, ".", 2)
	return spl[0], spl[1]
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
