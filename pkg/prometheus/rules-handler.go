package prometheus

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/mitchellh/mapstructure"
)

// RuleHandler is a Grizzly Provider for Grafana datasources
type RuleHandler struct{}

// NewRuleHandler returns configuration defining a new Grafana Provider
func NewRuleHandler() *RuleHandler {
	return &RuleHandler{}
}

// GetName returns the name for this provider
func (h *RuleHandler) GetName() string {
	return "prometheus"
}

// GetFullName returns the name for this provider
func (h *RuleHandler) GetFullName() string {
	return "prometheus.rulegroup"
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

// APIVersion returns the api version for this resource
func (h *RuleHandler) APIVersion() string {
	return "prometheus.io/v1"
}

// Kind returns the resource kind for this type of resource
func (h *RuleHandler) Kind() string {
	return "RuleGroup"
}
func (h *RuleHandler) newRuleGroupingResource(grouping RuleGrouping) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      grouping.Namespace,
		Filename: grouping.Namespace,
		Handler:  h,
		Detail:   grouping,
		JSONPath: "",
	}
	return resource
}

// ParseHiddenElements parses an interface{} object into a struct for this resource type
func (h *RuleHandler) ParseHiddenElements(path string, i interface{}) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	msi := i.(map[string]interface{})
	groupings := map[string]RuleGrouping{}
	err := mapstructure.Decode(msi, &groupings)
	if err != nil {
		return nil, err
	}
	for k, v := range groupings {
		v.Namespace = k
		m, err := grizzly.NewManifest(h, k, v)

		if err != nil {
			return nil, err
		}
		resource, err := h.Parse(m)
		if err != nil {
			return nil, err
		}
		resources[resource.Key()] = *resource
	}
	return resources, nil
}

// Parse parses a single resource from an interface{} object
func (h *RuleHandler) Parse(m manifest.Manifest) (*grizzly.Resource, error) {
	grouping := RuleGrouping{}
	grouping.Namespace = m.Metadata().Name()
	err := mapstructure.Decode(m["spec"], &grouping)
	if err != nil {
		return nil, err
	}
	resource := h.newRuleGroupingResource(grouping)
	return &resource, nil
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
	group, err := getRemoteRuleGrouping(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving rules %s: %v", UID, err)
	}
	resource := h.newRuleGroupingResource(*group)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (h *RuleHandler) GetRepresentation(uid string, resource grizzly.Resource) (string, error) {
	g := resource.Detail.(RuleGrouping)
	return g.toYAML()
}

// GetRemoteRepresentation retrieves a datasource as JSON
func (h *RuleHandler) GetRemoteRepresentation(uid string) (string, error) {
	group, err := getRemoteRuleGrouping(uid)
	if err != nil {
		return "", err
	}
	return group.toYAML()
}

// GetRemote retrieves a datasource as a Resource
func (h *RuleHandler) GetRemote(uid string) (*grizzly.Resource, error) {
	grouping, err := getRemoteRuleGrouping(uid)
	if err != nil {
		return nil, err
	}
	resource := h.newRuleGroupingResource(*grouping)
	return &resource, nil
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	g := resource.Detail.(RuleGrouping)
	return writeRuleGrouping(g)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	g := resource.Detail.(RuleGrouping)
	return writeRuleGrouping(g)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *RuleHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
	return grizzly.ErrNotImplemented
}
