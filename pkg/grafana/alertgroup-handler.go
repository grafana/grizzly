package grafana

import (
	"fmt"
	"path/filepath"
	"strings"

	"encoding/json"
	"errors"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// AlertRuleGroupHandler is a Grizzly Handler for Grafana alertRuleGroups
type AlertRuleGroupHandler struct {
	Provider grizzly.Provider
}

// NewAlertRuleGroupHandler returns a new Grizzly Handler for Grafana alertRuleGroups
func NewAlertRuleGroupHandler(provider grizzly.Provider) *AlertRuleGroupHandler {
	return &AlertRuleGroupHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *AlertRuleGroupHandler) Kind() string {
	return "AlertRuleGroup"
}

// Validate checks that the uid format is valid
func (h *AlertRuleGroupHandler) Validate(resource grizzly.Resource) error {
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var group models.AlertRuleGroup
	err = json.Unmarshal(data, &group)
	if err != nil {
		return err
	}
	uid := h.getUID(group)
	if uid != resource.Name() {
		return fmt.Errorf("title/folder combination '%s' and name '%s', don't match", uid, resource.Name())
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *AlertRuleGroupHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a alertRuleGroup
func (h *AlertRuleGroupHandler) GetExtension() string {
	return "json"
}

const (
	alertRuleGroupGlob    = "alert-rules/alertRuleGroup-*"
	alertRuleGroupPattern = "alert-rules/alertRuleGroup-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *AlertRuleGroupHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, alertRuleGroupGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertRuleGroupHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(alertRuleGroupPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *AlertRuleGroupHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *AlertRuleGroupHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *AlertRuleGroupHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *AlertRuleGroupHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertRuleGroupHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteAlertRuleGroup(UID)
}

// GetRemote retrieves a alertRuleGroup as a Resource
func (h *AlertRuleGroupHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteAlertRuleGroup(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertRuleGroupHandler) ListRemote() ([]string, error) {
	return h.getRemoteAlertRuleGroupList()
}

// Add pushes a alertRuleGroup to Grafana via the API
func (h *AlertRuleGroupHandler) Add(resource grizzly.Resource) error {
	return h.putAlertRuleGroup(resource)
}

// Update pushes a alertRuleGroup to Grafana via the API
func (h *AlertRuleGroupHandler) Update(existing, resource grizzly.Resource) error {
	return h.putAlertRuleGroup(resource)
}

// getRemoteAlertRuleGroup retrieves a alertRuleGroup object from Grafana
func (h *AlertRuleGroupHandler) getRemoteAlertRuleGroup(uid string) (*grizzly.Resource, error) {
	folder, group := h.splitUID(uid)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	alertRuleGroupOk, err := client.Provisioning.GetAlertRuleGroup(group, folder)
	if err != nil {
		var gErr *provisioning.GetAlertRuleGroupNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}
	alertRuleGroup := alertRuleGroupOk.GetPayload()
	// TODO: Turn spec into a real models.ProvisionedAlertRuleGroup object
	spec, err := structToMap(alertRuleGroup)
	if err != nil {
		return nil, err
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	resource.DeleteSpecKey("rules")
	return &resource, nil
}

func (h *AlertRuleGroupHandler) getRemoteAlertRuleGroupList() ([]string, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	alertRuleGroupsOk, err := client.Provisioning.GetAlertRules()
	if err != nil {
		return nil, err
	}
	alerts := alertRuleGroupsOk.GetPayload()

	uidmap := make(map[string]struct{})
	for _, alert := range alerts {
		uid := h.joinUID(*alert.FolderUID, *alert.RuleGroup)
		uidmap[uid] = struct{}{}
	}
	uids := make([]string, len(uidmap))
	idx := 0
	for k := range uidmap {
		uids[idx] = k
		idx++
	}
	return uids, nil
}

func (h *AlertRuleGroupHandler) putAlertRuleGroup(resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.AlertRuleGroup object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var group models.AlertRuleGroup
	err = json.Unmarshal(data, &group)

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	stringtrue := "true"
	params := provisioning.NewPutAlertRuleGroupParams().WithBody(&group).WithGroup(group.Title).WithFolderUID(group.FolderUID).WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutAlertRuleGroup(params, nil)
	return err
}

func (h *AlertRuleGroupHandler) getUID(group models.AlertRuleGroup) string {
	return h.joinUID(group.FolderUID, group.Title)
}

func (h *AlertRuleGroupHandler) joinUID(folder, title string) string {
	return fmt.Sprintf("%s.%s", folder, title)
}
func (h *AlertRuleGroupHandler) splitUID(uid string) (string, string) {
	spl := strings.SplitN(uid, ".", 2)
	return spl[0], spl[1]
}
