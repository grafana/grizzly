package grafana

import (
	"fmt"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	GlobalAlertNotificationPolicyName = "global"
)

// AlertNotificationPolicyHandler is a Grizzly Handler for Grafana alertNotificationPolicies
type AlertNotificationPolicyHandler struct {
	Provider grizzly.Provider
}

// NewAlertNotificationPolicyHandler returns a new Grizzly Handler for Grafana alertNotificationPolicies
func NewAlertNotificationPolicyHandler(provider grizzly.Provider) *AlertNotificationPolicyHandler {
	return &AlertNotificationPolicyHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *AlertNotificationPolicyHandler) Kind() string {
	return "AlertNotificationPolicy"
}

// Validate returns the uid of resource
func (h *AlertNotificationPolicyHandler) Validate(resource grizzly.Resource) error {
	if resource.Name() != GlobalAlertNotificationPolicyName {
		return fmt.Errorf("name of notification policy must be '%s', got '%s'", GlobalAlertNotificationPolicyName, resource.Name())
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *AlertNotificationPolicyHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

const (
	alertNotificationPolicyFile = "alertNotificationPolicy.yaml"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *AlertNotificationPolicyHandler) FindResourceFiles(dir string) ([]string, error) {
	p := filepath.Join(dir, alertNotificationPolicyFile)
	_, err := os.Stat(p)
	if err == nil {
		return []string{p}, nil
	}
	// just return empty, ignore error
	return nil, nil
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertNotificationPolicyHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return alertNotificationPolicyFile
}

// Parse parses a manifest object into a struct for this resource type
func (h *AlertNotificationPolicyHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, h.Validate(resource)
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *AlertNotificationPolicyHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *AlertNotificationPolicyHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *AlertNotificationPolicyHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertNotificationPolicyHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteAlertNotificationPolicy()
}

// GetRemote retrieves a alertNotificationPolicy as a Resource
func (h *AlertNotificationPolicyHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteAlertNotificationPolicy()
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertNotificationPolicyHandler) ListRemote() ([]string, error) {
	return h.getRemoteAlertNotificationPolicyList()
}

// Add pushes a alertNotificationPolicy to Grafana via the API
func (h *AlertNotificationPolicyHandler) Add(resource grizzly.Resource) error {
	return h.putAlertNotificationPolicy(resource)
}

// Update pushes a alertNotificationPolicy to Grafana via the API
func (h *AlertNotificationPolicyHandler) Update(existing, resource grizzly.Resource) error {
	return h.putAlertNotificationPolicy(resource)
}

// getRemoteAlertNotificationPolicy retrieves a alertNotificationPolicy object from Grafana
func (h *AlertNotificationPolicyHandler) getRemoteAlertNotificationPolicy() (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	resp, err := client.Provisioning.GetPolicyTree()
	if err != nil {
		return nil, err
	}

	policy := resp.GetPayload()

	// TODO: Turn spec into a real models.AlertNotificationPolicy object
	spec, err := structToMap(policy)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), GlobalAlertNotificationPolicyName, spec)
	return &resource, nil
}

func (h *AlertNotificationPolicyHandler) getRemoteAlertNotificationPolicyList() ([]string, error) {
	return []string{GlobalAlertNotificationPolicyName}, nil
}

func (h *AlertNotificationPolicyHandler) putAlertNotificationPolicy(resource grizzly.Resource) error {
	var alertNotificationPolicy models.Route
	// TODO: Turn spec into a real models.AlertNotificationPolicy object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &alertNotificationPolicy)
	if err != nil {
		return err
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	stringtrue := "true"
	params := provisioning.NewPutPolicyTreeParams().
		WithBody(&alertNotificationPolicy).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutPolicyTree(params)
	return err
}
