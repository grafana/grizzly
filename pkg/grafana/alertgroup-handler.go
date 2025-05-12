package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

const AlertRuleGroupKind = "AlertRuleGroup"

var _ grizzly.Handler = &AlertRuleGroupHandler{}
var _ grizzly.ProxyConfiguratorProvider = &AlertRuleGroupHandler{}

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

// ProxyConfigurator provides a configurator object describing how to proxy alert rule groups.
func (h *AlertRuleGroupHandler) ProxyConfigurator() grizzly.ProxyConfigurator {
	return &alertRuleProxyConfigurator{
		provider: h.Provider,
	}
}

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
		uid := joinAlertRuleGroupUID(*alert.FolderUID, *alert.RuleGroup)
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
	
	folderUID, ruleGroup := h.splitUID(resource.Name())
	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(&group).
		WithGroup(ruleGroup).
		WithFolderUID(folderUID).
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
	
	folderUID, ruleGroup := h.splitUID(resource.Name())
	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(group).
		WithGroup(ruleGroup).
		WithFolderUID(folderUID).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutAlertRuleGroup(params, nil)
	return err
}

func (h *AlertRuleGroupHandler) getUID(group models.AlertRuleGroup) string {
	return joinAlertRuleGroupUID(group.FolderUID, group.Title)
}

func joinAlertRuleGroupUID(folder, title string) string {
	return fmt.Sprintf("%s.%s", folder, title)
}

func (h *AlertRuleGroupHandler) splitUID(uid string) (string, string) {
	spl := strings.SplitN(uid, ".", 2)
	return spl[0], spl[1]
}
