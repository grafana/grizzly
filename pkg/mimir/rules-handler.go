package mimir

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly"
	"gopkg.in/yaml.v3"
)

// RuleHandler is a Grizzly Handler for Prometheus Rules
type RuleHandler struct {
	grizzly.BaseHandler
	cortexTool CortexTool
}

// NewRuleHandler returns a new Grizzly Handler for Prometheus Rules
func NewRuleHandler(provider *Provider) *RuleHandler {
	return &RuleHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "PrometheusRuleGroup", false),
		cortexTool:  NewCortexTool(provider.config),
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

	out, err := h.cortexTool.ExecuteCortexTool("rules", "print", "--disable-color")
	if err != nil {
		return nil, err
	}
	groupings := map[string][]PrometheusRuleGroup{}
	err = yaml.Unmarshal(out, &groupings)
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
	out, err := h.cortexTool.ExecuteCortexTool("rules", "print", "--disable-color")
	if err != nil {
		return nil, err
	}
	groupings := map[string][]PrometheusRuleGroup{}
	err = yaml.Unmarshal(out, &groupings)
	if err != nil {
		return nil, err
	}

	IDs := []string{}
	for namespace, grouping := range groupings {
		for _, group := range grouping {
			uid := fmt.Sprintf("%s.%s", namespace, group.Name)
			IDs = append(IDs, uid)
		}
	}
	return IDs, nil
}

// PrometheusRuleGroup encapsulates a list of rules
type PrometheusRuleGroup struct {
	Namespace string                   `yaml:"-"`
	Name      string                   `yaml:"name"`
	Rules     []map[string]interface{} `yaml:"rules"`
}

// PrometheusRuleGrouping encapsulates a set of named rule groups
type PrometheusRuleGrouping struct {
	Namespace string                `json:"namespace"`
	Groups    []PrometheusRuleGroup `json:"groups"`
}

func (h *RuleHandler) writeRuleGroup(resource grizzly.Resource) error {
	tmpfile, err := os.CreateTemp("", "cortextool-*")
	if err != nil {
		return err
	}
	newGroup := PrometheusRuleGroup{
		Name: resource.Name(),
		// Rules: resource.Spec()["rules"].([]map[string]interface{}),
		Rules: []map[string]interface{}{},
	}
	rules := resource.Spec()["rules"].([]interface{})
	for _, ruleIf := range rules {
		rule := ruleIf.(map[string]interface{})
		newGroup.Rules = append(newGroup.Rules, rule)
	}
	grouping := PrometheusRuleGrouping{
		Namespace: resource.GetMetadata("namespace"),
		Groups:    []PrometheusRuleGroup{newGroup},
	}
	out, err := yaml.Marshal(grouping)
	if err != nil {
		return err
	}
	os.WriteFile(tmpfile.Name(), out, 0644)

	output, err := h.cortexTool.ExecuteCortexTool("rules", "load", tmpfile.Name())
	if err != nil {
		log.Println(output)
		return err
	}
	os.Remove(tmpfile.Name())
	return err
}
