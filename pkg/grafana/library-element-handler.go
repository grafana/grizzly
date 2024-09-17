package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	library "github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
	log "github.com/sirupsen/logrus"
)

const LibraryElementKind = "LibraryElement"

// LibraryElementHandler is a Grizzly Handler for Grafana dashboard folders
type LibraryElementHandler struct {
	grizzly.BaseHandler
}

var _ grizzly.Handler = &LibraryElementHandler{}

// NewLibraryElementHandler returns configuration defining a new Grafana Library Element Handler
func NewLibraryElementHandler(provider grizzly.Provider) *LibraryElementHandler {
	return &LibraryElementHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, LibraryElementKind, false),
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

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *LibraryElementHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("meta")
	resource.DeleteSpecKey("version")
	resource.DeleteSpecKey("id")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *LibraryElementHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	if existing != nil {
		val := existing.GetSpecValue("version")
		resource.SetSpecValue("version", val)
	}
	resource.DeleteSpecKey("meta")

	uid, _ := resource.GetSpecString("uid")
	if uid == "" {
		resource.SetSpecString("uid", resource.Name())
	}
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
	uid, ok := resource.GetSpecString("uid")
	if !ok {
		return "", fmt.Errorf("UID not specified")
	}
	return uid, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *LibraryElementHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	resource, err := h.getRemoteLibraryElement(uid)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving library element %s: %w", uid, err)
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

func (h *LibraryElementHandler) GetProxyEndpoints(s grizzly.Server) []grizzly.HTTPEndpoint {
	return []grizzly.HTTPEndpoint{
		{
			Method:  "GET",
			URL:     "/api/library-elements/{uid}",
			Handler: h.LibraryElementJSONGetHandler(s),
		},
	}
}

func (h *LibraryElementHandler) ProxyURL(uid string) string {
	return fmt.Sprintf("/api/library-elements/%s", uid)
}

func (h *LibraryElementHandler) LibraryElementJSONGetHandler(s grizzly.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		if uid == "" {
			grizzly.SendError(w, "No UID specified", fmt.Errorf("no UID specified within the URL"), 400)
			return
		}

		resource, found := s.Resources.Find(grizzly.NewResourceRef(LibraryElementKind, uid))
		if !found {
			log.Debug("Library element not found in memory, proxying request...", "uid", uid)
			s.ProxyRequestHandler(w, r)
			return
		}

		if resource.GetSpecValue("version") == nil {
			resource.SetSpecValue("version", 1)
		}

		writeJSONOrLog(w, map[string]any{
			"result": resource.Spec(),
		})
	}
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
