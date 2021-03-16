package grafana

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/mitchellh/mapstructure"
)

// RuleHandler is a Grizzly Handler for Prometheus Rules
type RuleHandler struct {
	Provider Provider
}

// NewRuleHandler returns a new Grizzly Handler for Prometheus Rules
func NewRuleHandler(provider Provider) *RuleHandler {
	return &RuleHandler{
		Provider: provider,
	}
}

// Kind returns the name for this handler
func (h *RuleHandler) Kind() string {
	return "PrometheusRuleGroup"
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *RuleHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

const prometheusAlertsPath = "prometheusAlerts"
const prometheusRulesPath = "prometheusRules"

// GetJSONPaths returns paths within Jsonnet output that this provider will consume
func (h *RuleHandler) GetJSONPaths() []string {
	return []string{
		prometheusAlertsPath,
		prometheusRulesPath,
	}
}

// GetExtension returns the file name extension for a rule grouping
func (h *RuleHandler) GetExtension() string {
	return "yaml"
}

func (h *RuleHandler) newRuleGroupResource(path string, group PrometheusRuleGroup) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      group.UID(),
		Filename: group.UID(),
		Handler:  h,
		Detail:   group,
		JSONPath: path,
	}
	return resource
}

// Parse parses a manifest object into a struct for this resource type
func (h *RuleHandler) Parse(m manifest.Manifest) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	spec := m["spec"].(map[string]interface{})
	group := PrometheusRuleGroup{}
	err := mapstructure.Decode(spec, &group)
	if err != nil {
		return nil, err
	}
	group.Namespace = m.Metadata().Namespace()
	group.Name = m.Metadata().Name()
	resource := h.newRuleGroupResource("", group)
	key := resource.Key()
	resources[key] = resource
	return resources, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *RuleHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *RuleHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *RuleHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	group, err := getRemoteRuleGroup(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving datasource %s: %v", UID, err)
	}
	resource := h.newRuleGroupResource(prometheusAlertsPath, *group)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (h *RuleHandler) GetRepresentation(uid string, resource grizzly.Resource) (string, error) {
	g := resource.Detail.(PrometheusRuleGroup)
	return g.toYAML()
}

// GetRemoteRepresentation retrieves a datasource as JSON
func (h *RuleHandler) GetRemoteRepresentation(uid string) (string, error) {
	group, err := getRemoteRuleGroup(uid)
	if err != nil {
		return "", err
	}
	return group.toYAML()
}

// GetRemote retrieves a datasource as a Resource
func (h *RuleHandler) GetRemote(uid string) (*grizzly.Resource, error) {
	group, err := getRemoteRuleGroup(uid)
	if err != nil {
		return nil, err
	}
	resource := h.newRuleGroupResource("", *group)
	return &resource, nil
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	g := resource.Detail.(PrometheusRuleGroup)
	return writeRuleGroup(g)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	g := resource.Detail.(PrometheusRuleGroup)
	return writeRuleGroup(g)
}
