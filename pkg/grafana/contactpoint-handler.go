package grafana

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/internal/utils"
	"github.com/grafana/grizzly/pkg/grizzly"
	log "github.com/sirupsen/logrus"
)

const AlertContactPointKind = "AlertContactPoint"

var _ grizzly.Handler = &AlertContactPointHandler{}

// AlertContactPointHandler is a Grizzly Handler for Grafana contactPoints
type AlertContactPointHandler struct {
	grizzly.BaseHandler
}

// NewAlertContactPointHandler returns a new Grizzly Handler for Grafana contactPoints
func NewAlertContactPointHandler(provider grizzly.Provider) *AlertContactPointHandler {
	return &AlertContactPointHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, AlertContactPointKind, false),
	}
}

const (
	contactPointPattern = "alert-contact-points/contactPoint-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertContactPointHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(contactPointPattern, resource.Name(), filetype)
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *AlertContactPointHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	uid, _ := resource.GetSpecString("uid")
	if uid == "" {
		resource.SetSpecString("uid", resource.Name())
	}
	return &resource
}

// Validate returns the uid of resource
func (h *AlertContactPointHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist && uid != resource.Name() {
		return ErrUIDNameMismatch{UID: uid, Name: resource.Name()}
	}
	return nil
}

func (h *AlertContactPointHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	uid, ok := resource.GetSpecString("uid")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}
	return uid, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertContactPointHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return h.getRemoteContactPoint(uid)
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
	resource, err := h.getRemoteContactPointWithDecrypt(uid)
	if err != nil {
		var gErr *provisioning.GetContactpointsExportForbidden
		if !errors.As(err, &gErr) {
			return nil, err
		}
	}
	if err == nil {
		return resource, nil
	}

	log.Warn("Insufficient permissions to decrypt secrets in contact points. An admin service account is needed for that action, falling back to redacted secrets instead.")

	return h.getRemoteContactPointFallbackNoDecrypt(uid)
}

func (h *AlertContactPointHandler) getRemoteContactPointWithDecrypt(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	params := provisioning.NewGetContactpointsExportParams()
	params.Decrypt = utils.ToPtr(true)
	contactPoints, err := client.Provisioning.GetContactpointsExport(params)
	if err != nil {
		return nil, err
	}

	var point *models.ContactPointExport
	var receiver *models.ReceiverExport
	for _, c := range contactPoints.GetPayload().ContactPoints {
		for _, r := range c.Receivers {
			if r.UID == uid {
				point = c
				receiver = r
				break
			}
		}
		if point != nil {
			break
		}
	}
	if point == nil || receiver == nil {
		return nil, grizzly.ErrNotFound
	}

	// TODO: Turn spec into a real models.ContactPoint object
	spec := map[string]any{
		"name":                  point.Name,
		"uid":                   receiver.UID,
		"type":                  receiver.Type,
		"settings":              receiver.Settings,
		"disableResolveMessage": receiver.DisableResolveMessage,
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

func (h *AlertContactPointHandler) getRemoteContactPointFallbackNoDecrypt(uid string) (*grizzly.Resource, error) {
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
	params := provisioning.NewPutContactpointParams().
		WithUID(resource.Name()).
		WithBody(&modelContactPoint).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutContactpoint(params)
	return err
}
