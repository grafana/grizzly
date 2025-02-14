package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

const (
	GlobalAlertNotificationPolicyName = "global"
	AlertNotificationPolicyKind       = "AlertNotificationPolicy"
)

var _ grizzly.Handler = &AlertNotificationPolicyHandler{}

// AlertNotificationPolicyHandler is a Grizzly Handler for Grafana alertNotificationPolicies
type AlertNotificationPolicyHandler struct {
	grizzly.BaseHandler
}

// NewAlertNotificationPolicyHandler returns a new Grizzly Handler for Grafana alertNotificationPolicies
func NewAlertNotificationPolicyHandler(provider grizzly.Provider) *AlertNotificationPolicyHandler {
	return &AlertNotificationPolicyHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, AlertNotificationPolicyKind, false),
	}
}

const (
	alertNotificationPolicyFile = "alertNotificationPolicy.yaml"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertNotificationPolicyHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return alertNotificationPolicyFile
}

// Validate returns the uid of resource
func (h *AlertNotificationPolicyHandler) Validate(resource grizzly.Resource) error {
	if resource.Name() != GlobalAlertNotificationPolicyName {
		return fmt.Errorf("name of notification policy must be '%s', got '%s'", GlobalAlertNotificationPolicyName, resource.Name())
	}
	return nil
}

func (h *AlertNotificationPolicyHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	return AlertNotificationPolicyKind + "-UID", nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertNotificationPolicyHandler) GetByUID(uid string) (*grizzly.Resource, error) {
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

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), GlobalAlertNotificationPolicyName, spec)
	if err != nil {
		return nil, err
	}
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
	params := provisioning.NewPutPolicyTreeParams().
		WithBody(&alertNotificationPolicy).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutPolicyTree(params)
	return err
}
