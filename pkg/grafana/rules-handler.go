package grafana

import (
	"fmt"
	"path/filepath"

	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"gopkg.in/yaml.v3"
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

// Validate returns the uid of resource
func (h *RuleHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
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

// GetUID returns the UID for a resource
func (h *RuleHandler) GetUID(resource grizzly.Resource) (string, error) {
	if !resource.HasMetadata("namespace") {
		return "", fmt.Errorf("%s %s requires a namespace metadata entry", h.Kind(), resource.Name())
	}
	return fmt.Sprintf("%s.%s", resource.GetMetadata("namespace"), resource.Name()), nil
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

var cortexTool = func(args ...string) ([]byte, error) {
	path := os.Getenv("CORTEXTOOL_PATH")
	if path == "" {
		var err error
		path, err = exec.LookPath("cortextool")
		if err != nil {
			return nil, err
		} else if path == "" {
			return nil, fmt.Errorf("cortextool not found")
		}
	}
	return exec.Command(path, args...).Output()
}

// getRemoteRuleGroup retrieves a datasource object from Grafana
func (h *RuleHandler) getRemoteRuleGroup(uid string) (*grizzly.Resource, error) {
	parts := strings.SplitN(uid, ".", 2)
	namespace := parts[0]
	name := parts[1]
	out, err := cortexTool("rules", "print", "--disable-color")
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
					handler := RuleHandler{}
					resource := grizzly.NewResource(handler.APIVersion(), handler.Kind(), group.Name, spec)
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
	out, err := cortexTool("rules", "print", "--disable-color")
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
		//Rules: resource.Spec()["rules"].([]map[string]interface{}),
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
	output, err := cortexTool("rules", "load", tmpfile.Name())
	if err != nil {
		log.Println("OUTPUT", output)
		return err
	}
	os.Remove(tmpfile.Name())
	return err
}
