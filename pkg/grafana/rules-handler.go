package grafana

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
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

// Parse parses a manifest object into a struct for this resource type
func (h *RuleHandler) Parse(m manifest.Manifest) (grizzly.ResourceList, error) {
	resource := grizzly.Resource(m)
	return resource.AsResourceList(), nil
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
	return getRemoteRuleGroup(UID)
}

// GetRemote retrieves a datasource as a Resource
func (h *RuleHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid := fmt.Sprintf("%s.%s", resource.GetMetadata("namespace"), resource.Name())
	return getRemoteRuleGroup(uid)
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	return writeRuleGroup(resource)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	return writeRuleGroup(resource)
}
