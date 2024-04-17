package mimir

import (
	"fmt"
	"github.com/grafana/grizzly/pkg/mimir/client"
	"github.com/grafana/grizzly/pkg/mimir/models"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly"
)

// RuleHandler is a Grizzly Handler for Prometheus Rules
type RuleHandler struct {
	grizzly.BaseHandler
	clientTool client.Mimir
}

// NewRuleHandler returns a new Grizzly Handler for Prometheus Rules
func NewRuleHandler(provider *Provider, clientTool client.Mimir) *RuleHandler {
	return &RuleHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "PrometheusRuleGroup", false),
		clientTool:  clientTool,
	}
}

const (
	prometheusRuleGroupPattern = "prometheus/rules-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *RuleHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(prometheusRuleGroupPattern, resource.Name(), filetype)
}

// Validate returns the uid of resource
func (h *RuleHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist && uid != resource.Name() {
		return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return nil
}

// GetUID returns the UID for a resource
func (h *RuleHandler) GetUID(resource grizzly.Resource) (string, error) {
	if !resource.HasMetadata("namespace") {
		return "", fmt.Errorf("%s %s requires a namespace metadata entry", h.Kind(), resource.Name())
	}
	return fmt.Sprintf("%s.%s", resource.GetMetadata("namespace"), resource.Name()), nil
}

func (h *RuleHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	return "", fmt.Errorf("GetSpecUID not implemented for prometheus rules")
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *RuleHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteRuleGroup(UID)
}

// GetRemote retrieves a datasource as a Resource
func (h *RuleHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid := fmt.Sprintf("%s.%s", resource.GetMetadata("namespace"), resource.Name())
	return h.getRemoteRuleGroup(uid)
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *RuleHandler) ListRemote() ([]string, error) {
	return h.getRemoteRuleGroupList()
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	return h.writeRuleGroup(resource)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	return h.writeRuleGroup(resource)
}

// getRemoteRuleGroup retrieves a datasource object from Grafana
func (h *RuleHandler) getRemoteRuleGroup(uid string) (*grizzly.Resource, error) {
	parts := strings.SplitN(uid, ".", 2)
	namespace := parts[0]
	name := parts[1]

	groupings, err := h.clientTool.ListRules()
	if err != nil {
		return nil, err
	}

	for key, grouping := range groupings {
		if key == namespace {
			for _, group := range grouping {
				if group.Name == name {
					spec := map[string]interface{}{
						"rules": group.Rules,
					}
					resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), group.Name, spec)
					if err != nil {
						return nil, err
					}
					resource.SetMetadata("namespace", namespace)
					return &resource, nil
				}
			}
		}
	}
	return nil, grizzly.ErrNotFound
}

// getRemoteRuleGroupList retrieves a datasource object from Grafana
func (h *RuleHandler) getRemoteRuleGroupList() ([]string, error) {
	groupings, err := h.clientTool.ListRules()
	if err != nil {
		return nil, err
	}

	var IDs []string
	for namespace, grouping := range groupings {
		for _, group := range grouping {
			uid := fmt.Sprintf("%s.%s", namespace, group.Name)
			IDs = append(IDs, uid)
		}
	}
	return IDs, nil
}

func (h *RuleHandler) writeRuleGroup(resource grizzly.Resource) error {
	newGroup := models.PrometheusRuleGroup{
		Name:  resource.Name(),
		Rules: []interface{}{},
	}
	rules := resource.Spec()["rules"].([]interface{})
	for _, ruleIf := range rules {
		rule := ruleIf.(map[string]interface{})
		newGroup.Rules = append(newGroup.Rules, rule)
	}
	grouping := models.PrometheusRuleGrouping{
		Namespace: resource.GetMetadata("namespace"),
		Groups:    []models.PrometheusRuleGroup{newGroup},
	}

	return h.clientTool.CreateRules(grouping)
}
