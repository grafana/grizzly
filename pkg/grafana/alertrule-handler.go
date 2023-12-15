package grafana

import (
	"fmt"
	"path/filepath"

	"encoding/json"
	"errors"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// AlertRuleHandler is a Grizzly Handler for Grafana alertRules
type AlertRuleHandler struct {
	Provider grizzly.Provider
}

// NewAlertRuleHandler returns a new Grizzly Handler for Grafana alertRules
func NewAlertRuleHandler(provider grizzly.Provider) *AlertRuleHandler {
	return &AlertRuleHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *AlertRuleHandler) Kind() string {
	return "AlertRule"
}

// Validate returns the uid of resource
func (h *AlertRuleHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *AlertRuleHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a alertRule
func (h *AlertRuleHandler) GetExtension() string {
	return "json"
}

const (
	alertRuleGlob    = "alert-rules/alertRule-*"
	alertRulePattern = "alert-rules/alertRule-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *AlertRuleHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, alertRuleGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertRuleHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(alertRulePattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *AlertRuleHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *AlertRuleHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *AlertRuleHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *AlertRuleHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertRuleHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteAlertRule(UID)
}

// GetRemote retrieves a alertRule as a Resource
func (h *AlertRuleHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteAlertRule(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertRuleHandler) ListRemote() ([]string, error) {
	return h.getRemoteAlertRuleList()
}

// Add pushes a alertRule to Grafana via the API
func (h *AlertRuleHandler) Add(resource grizzly.Resource) error {
	return h.postAlertRule(resource)
}

// Update pushes a alertRule to Grafana via the API
func (h *AlertRuleHandler) Update(existing, resource grizzly.Resource) error {
	return h.putAlertRule(resource)
}

// getRemoteAlertRule retrieves a alertRule object from Grafana
func (h *AlertRuleHandler) getRemoteAlertRule(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	alertRuleOk, err := client.Provisioning.GetAlertRule(uid)
	if err != nil {
		var gErr *provisioning.GetAlertRuleNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}
	alertRule := alertRuleOk.GetPayload()
	// TODO: Turn spec into a real models.ProvisionedAlertRule object
	spec, err := structToMap(alertRule)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	return &resource, nil
}

func (h *AlertRuleHandler) getRemoteAlertRuleList() ([]string, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	alertRulesOk, err := client.Provisioning.GetAlertRules()
	if err != nil {
		return nil, err
	}
	alertRules := alertRulesOk.GetPayload()

	uids := make([]string, len(alertRules))
	for i, alertRule := range alertRules {
		uids[i] = alertRule.UID
	}
	return uids, nil
}

func (h *AlertRuleHandler) postAlertRule(resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.AlertRule object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var rule models.ProvisionedAlertRule
	err = json.Unmarshal(data, &rule)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	stringtrue := "true"
	params := provisioning.NewPostAlertRuleParams().WithBody(&rule).WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PostAlertRule(params, nil)
	return err
}

func (h *AlertRuleHandler) putAlertRule(resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.AlertRule object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var rule models.ProvisionedAlertRule
	err = json.Unmarshal(data, &rule)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	stringtrue := "true"
	params := provisioning.NewPutAlertRuleParams().
		WithUID(resource.UID()).
		WithBody(&rule).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutAlertRule(params, nil)
	return err
}
