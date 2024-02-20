package grafana

import (
	"encoding/json"
	"fmt"

	"errors"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"

	library "github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// LibraryElementHandler is a Grizzly Handler for Grafana dashboard folders
type LibraryElementHandler struct {
	grizzly.BaseHandler
}

var _ grizzly.Handler = &LibraryElementHandler{}

// NewLibraryElementHandler returns configuration defining a new Grafana Library Element Handler
func NewLibraryElementHandler(provider grizzly.Provider) *LibraryElementHandler {
	return &LibraryElementHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "LibraryElement", false),
	}
}

const (
	libraryElementPattern = "library-elements/%s-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *LibraryElementHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	kind := "element"
	t := resource.GetSpecValue("kind").(float64)

	switch t {
	case 1:
		kind = "panel"
	case 2:
		kind = "variable"
	}
	return fmt.Sprintf(libraryElementPattern, kind, resource.Name(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *LibraryElementHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource, err := grizzly.ResourceFromMap(m)
	if err != nil {
		return nil, err
	}
	resource.SetSpecString("uid", resource.Name())
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *LibraryElementHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("meta")
	resource.DeleteSpecKey("version")
	resource.DeleteSpecKey("id")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *LibraryElementHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	if existing != nil {
		val := existing.GetSpecValue("version")
		resource.SetSpecValue("version", val)
	}
	resource.DeleteSpecKey("meta")
	return &resource
}

// Validate returns the uid of resource
func (h *LibraryElementHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist {
		if uid != resource.Name() {
			return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
		}
	}

	return nil
}

func (h *LibraryElementHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	spec := resource["spec"].(map[string]interface{})
	if val, ok := spec["uid"]; ok {
		return val.(string), nil
	}
	return "", fmt.Errorf("UID not specified")
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *LibraryElementHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteLibraryElement(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving library element %s: %w", UID, err)
	}

	return resource, nil
}

// GetRemote retrieves an element as a resource
func (h *LibraryElementHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteLibraryElement(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *LibraryElementHandler) ListRemote() ([]string, error) {
	return h.listElements()
}

// Add pushes a new element to Grafana via the API
func (h *LibraryElementHandler) Add(resource grizzly.Resource) error {
	return h.createElement(resource)
}

// Update pushes an element to Grafana via the API
func (h *LibraryElementHandler) Update(existing, resource grizzly.Resource) error {
	return h.updateElement(existing, resource)
}

func (h *LibraryElementHandler) listElements() ([]string, error) {
	params := library.NewGetLibraryElementsParams()
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	elemsOK, err := client.LibraryElements.GetLibraryElements(params, nil)
	if err != nil {
		return nil, err
	}
	elems := elemsOK.GetPayload().Result.Elements
	uids := make([]string, len(elems))
	for i, e := range elems {
		uids[i] = e.UID
	}
	return uids, nil
}

func (h *LibraryElementHandler) updateElement(existing, resource grizzly.Resource) error {
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var command models.PatchLibraryElementCommand
	err = json.Unmarshal(data, &command)
	if err != nil {
		return err
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	_, err = client.LibraryElements.UpdateLibraryElement(resource.Name(), &command)
	return err
}

func (h *LibraryElementHandler) createElement(resource grizzly.Resource) error {
	data, err := json.Marshal(resource.Spec())
	if err != nil {
		return err
	}
	var command models.CreateLibraryElementCommand
	err = json.Unmarshal(data, &command)
	if err != nil {
		return err
	}
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	_, err = client.LibraryElements.CreateLibraryElement(&command, nil)
	return err
}

func (h *LibraryElementHandler) getRemoteLibraryElement(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	libraryElementsOk, err := client.LibraryElements.GetLibraryElementByUID(uid, nil)
	if err != nil {
		var gErr *library.GetLibraryElementByUIDNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}
	libraryElement := libraryElementsOk.GetPayload()

	spec, err := structToMap(libraryElement.Result)
	if err != nil {
		return nil, err
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}
