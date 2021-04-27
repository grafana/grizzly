package grafana

import (
	"fmt"
	"path/filepath"

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

// GetExtension returns the file name extension for a rule grouping
func (h *RuleHandler) GetExtension() string {
	return "yaml"
}

const (
	prometheusRuleGroupGlob    = "prometheus/rules-*"
	prometheusRuleGroupPattern = "prometheus/rules-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *RuleHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, prometheusRuleGroupGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *RuleHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(prometheusRuleGroupPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *RuleHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
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

// ListRemote retrieves as list of UIDs of all remote resources
func (h *RuleHandler) ListRemote() ([]string, error) {
	return getRemoteRuleGroupList()
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	return writeRuleGroup(resource)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	return writeRuleGroup(resource)
}
