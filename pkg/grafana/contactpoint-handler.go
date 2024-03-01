package grafana

import (
	"fmt"

	"encoding/json"

	"github.com/grafana/grizzly/pkg/grizzly"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// AlertContactPointHandler is a Grizzly Handler for Grafana contactPoints
type AlertContactPointHandler struct {
	grizzly.BaseHandler
}

// NewAlertContactPointHandler returns a new Grizzly Handler for Grafana contactPoints
func NewAlertContactPointHandler(provider grizzly.Provider) *AlertContactPointHandler {
	return &AlertContactPointHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "AlertContactPoint", false),
	}
}

const (
	contactPointPattern = "alert-contact-points/contactPoint-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertContactPointHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(contactPointPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *AlertContactPointHandler) Parse(m map[string]any) (grizzly.Resource, error) {
	resource, err := grizzly.ResourceFromMap(m)
	if err != nil {
		return nil, err
	}
	resource.SetSpecString("uid", resource.Name())
	return resource, nil
}

// Validate returns the uid of resource
func (h *AlertContactPointHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

func (h *AlertContactPointHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	spec := resource["spec"].(map[string]interface{})
	if val, ok := spec["uid"]; ok {
		return val.(string), nil
	}
	return "", fmt.Errorf("UID not specified")
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertContactPointHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteContactPoint(UID)
}

// GetRemote retrieves a contactPoint as a Resource
func (h *AlertContactPointHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteContactPoint(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertContactPointHandler) ListRemote() ([]string, error) {
	return h.getRemoteContactPointList()
}

// Add pushes a contactPoint to Grafana via the API
func (h *AlertContactPointHandler) Add(resource grizzly.Resource) error {
	return h.postContactPoint(resource)
}

// Update pushes a contactPoint to Grafana via the API
func (h *AlertContactPointHandler) Update(existing, resource grizzly.Resource) error {
	return h.putContactPoint(resource)
}

// getRemoteContactPoint retrieves a contactPoint object from Grafana
func (h *AlertContactPointHandler) getRemoteContactPoint(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	params := provisioning.NewGetContactpointsParams()
	contactPoints, err := client.Provisioning.GetContactpoints(params)
	if err != nil {
		return nil, err
	}
	var point *models.EmbeddedContactPoint
	for _, c := range contactPoints.GetPayload() {
		if c.UID == uid {
			point = c
			break
		}
	}
	if point == nil {
		return nil, grizzly.ErrNotFound
	}

	// TODO: Turn spec into a real models.ContactPoint object
	spec, err := structToMap(point)
	if err != nil {
		return nil, err
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (h *AlertContactPointHandler) getRemoteContactPointList() ([]string, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	params := provisioning.NewGetContactpointsParams()
	contactPointsOk, err := client.Provisioning.GetContactpoints(params)
	if err != nil {
		return nil, err
	}
	contactPoints := contactPointsOk.GetPayload()
	uids := make([]string, len(contactPoints))
	for i, contactPoint := range contactPoints {
		uids[i] = contactPoint.UID
	}
	return uids, nil
}

func (h *AlertContactPointHandler) postContactPoint(resource grizzly.Resource) error {
	var contactPoint models.EmbeddedContactPoint
	// TODO: Turn spec into a real models.ContactPoint object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &contactPoint)
	if err != nil {
		return err
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	stringtrue := "true"
	params := provisioning.NewPostContactpointsParams().
		WithBody(&contactPoint).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PostContactpoints(params)
	return err
}

func (h *AlertContactPointHandler) putContactPoint(resource grizzly.Resource) error {
	// TODO: Turn spec into a real models.ContactPoint object
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}

	var modelContactPoint models.EmbeddedContactPoint
	err = json.Unmarshal(data, &modelContactPoint)
	if err != nil {
		return err
	}

	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	stringtrue := "true"
	params := provisioning.NewPutContactpointParams().
		WithUID(resource.Name()).
		WithBody(&modelContactPoint).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutContactpoint(params)
	return err
}
