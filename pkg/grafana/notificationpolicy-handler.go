package grafana

import (
	"fmt"
	"path/filepath"

	"encoding/json"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	GlobalNotificationPolicyName = "global"
)

// NotificationPolicyHandler is a Grizzly Handler for Grafana notificationPolicys
type NotificationPolicyHandler struct {
	Provider grizzly.Provider
}

// NewNotificationPolicyHandler returns a new Grizzly Handler for Grafana notificationPolicys
func NewNotificationPolicyHandler(provider grizzly.Provider) *NotificationPolicyHandler {
	return &NotificationPolicyHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *NotificationPolicyHandler) Kind() string {
	return "NotificationPolicy"
}

// Validate returns the uid of resource
func (h *NotificationPolicyHandler) Validate(resource grizzly.Resource) error {
	if resource.Name() != GlobalNotificationPolicyName {
		return fmt.Errorf("name of notification policy must be '%s', got '%s'", GlobalNotificationPolicyName, resource.Name())
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *NotificationPolicyHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a notificationPolicy
func (h *NotificationPolicyHandler) GetExtension() string {
	return "json"
}

const (
	notificationPolicyFile = "notificationpolicy.yaml"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *NotificationPolicyHandler) FindResourceFiles(dir string) ([]string, error) {
	return []string{filepath.Join(dir, notificationPolicyFile)}, nil
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *NotificationPolicyHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return notificationPolicyFile
}

// Parse parses a manifest object into a struct for this resource type
func (h *NotificationPolicyHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, h.Validate(resource)
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *NotificationPolicyHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *NotificationPolicyHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *NotificationPolicyHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *NotificationPolicyHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteNotificationPolicy()
}

// GetRemote retrieves a notificationPolicy as a Resource
func (h *NotificationPolicyHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteNotificationPolicy()
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *NotificationPolicyHandler) ListRemote() ([]string, error) {
	return h.getRemoteNotificationPolicyList()
}

// Add pushes a notificationPolicy to Grafana via the API
func (h *NotificationPolicyHandler) Add(resource grizzly.Resource) error {
	return h.putNotificationPolicy(resource)
}

// Update pushes a notificationPolicy to Grafana via the API
func (h *NotificationPolicyHandler) Update(existing, resource grizzly.Resource) error {
	return h.putNotificationPolicy(resource)
}

// getRemoteNotificationPolicy retrieves a notificationPolicy object from Grafana
func (h *NotificationPolicyHandler) getRemoteNotificationPolicy() (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	resp, err := client.Provisioning.GetPolicyTree()
	if err != nil {
		return nil, err
	}

	policy := resp.GetPayload()

	// TODO: Turn spec into a real models.NotificationPolicy object
	spec, err := structToMap(policy)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), GlobalNotificationPolicyName, spec)
	return &resource, nil
}

func (h *NotificationPolicyHandler) getRemoteNotificationPolicyList() ([]string, error) {
	return []string{GlobalNotificationPolicyName}, nil
}

func (h *NotificationPolicyHandler) putNotificationPolicy(resource grizzly.Resource) error {
	var notificationPolicy models.Route
	// TODO: Turn spec into a real models.NotificationPolicy object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &notificationPolicy)
	if err != nil {
		return err
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	stringtrue := "true"
	params := provisioning.NewPutPolicyTreeParams().
		WithBody(&notificationPolicy).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutPolicyTree(params)
	return err
}
