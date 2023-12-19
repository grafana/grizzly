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

// ContactPointHandler is a Grizzly Handler for Grafana contactPoints
type ContactPointHandler struct {
	Provider grizzly.Provider
}

// NewContactPointHandler returns a new Grizzly Handler for Grafana contactPoints
func NewContactPointHandler(provider grizzly.Provider) *ContactPointHandler {
	return &ContactPointHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *ContactPointHandler) Kind() string {
	return "ContactPoint"
}

// Validate returns the uid of resource
func (h *ContactPointHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *ContactPointHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a contactPoint
func (h *ContactPointHandler) GetExtension() string {
	return "json"
}

const (
	contactPointGlob    = "contact-points/contactPoint-*"
	contactPointPattern = "contact-points/contactPoint-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *ContactPointHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, contactPointGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *ContactPointHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(contactPointPattern, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *ContactPointHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	resource.SetSpecString("uid", resource.Name())
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *ContactPointHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *ContactPointHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *ContactPointHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *ContactPointHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return h.getRemoteContactPoint(UID)
}

// GetRemote retrieves a contactPoint as a Resource
func (h *ContactPointHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteContactPoint(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *ContactPointHandler) ListRemote() ([]string, error) {
	return h.getRemoteContactPointList()
}

// Add pushes a contactPoint to Grafana via the API
func (h *ContactPointHandler) Add(resource grizzly.Resource) error {
	return h.postContactPoint(resource)
}

// Update pushes a contactPoint to Grafana via the API
func (h *ContactPointHandler) Update(existing, resource grizzly.Resource) error {
	return h.putContactPoint(resource)
}

// getRemoteContactPoint retrieves a contactPoint object from Grafana
func (h *ContactPointHandler) getRemoteContactPoint(uid string) (*grizzly.Resource, error) {
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

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	return &resource, nil
}

func (h *ContactPointHandler) getRemoteContactPointList() ([]string, error) {
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

func (h *ContactPointHandler) postContactPoint(resource grizzly.Resource) error {
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

func (h *ContactPointHandler) putContactPoint(resource grizzly.Resource) error {
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
		WithUID(resource.UID()).
		WithBody(&modelContactPoint).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutContactpoint(params)
	return err
}
