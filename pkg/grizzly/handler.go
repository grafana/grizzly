package grizzly

import (
	"net/http"
)

type BaseHandler struct {
	Provider    Provider
	kind        string
	usesFolders bool
}

func NewBaseHandler(provider Provider, kind string, usesFolders bool) BaseHandler {
	return BaseHandler{
		Provider:    provider,
		kind:        kind,
		usesFolders: usesFolders,
	}
}

func (h *BaseHandler) Kind() string {
	return h.kind
}

func (h *BaseHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

func (h *BaseHandler) UsesFolders() bool {
	return h.usesFolders
}

func (h *BaseHandler) Unprepare(resource Resource) *Resource {
	return &resource
}

func (h *BaseHandler) Prepare(existing, resource Resource) *Resource {
	return &resource
}

func (h *BaseHandler) GetUID(resource Resource) (string, error) {
	return resource.Name(), nil
}

func (h *BaseHandler) Sort(resources Resources) Resources {
	return resources
}

func (h *BaseHandler) Detect(map[string]any) bool {
	return false
}

// Handler describes a handler for a single API resource handled by a single provider
type Handler interface {
	APIVersion() string
	Kind() string

	// ResourceFilePath returns the location on disk where a resource should be updated
	ResourceFilePath(resource Resource, filetype string) string

	// Parse parses a manifest object into a struct for this resource type
	Parse(m map[string]any) (Resource, error)

	// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
	Unprepare(resource Resource) *Resource

	// Prepare gets a resource ready for dispatch to the remote endpoint
	Prepare(existing, resource Resource) *Resource

	// Retrieves a UID for a resource
	GetUID(resource Resource) (string, error)

	// GetSpecUID retrieves a UID from the spec of a raw resource
	GetSpecUID(resource Resource) (string, error)

	// Get retrieves JSON for a resource from an endpoint, by UID
	GetByUID(UID string) (*Resource, error)

	// GetRemote retrieves a remote equivalent of a remote resource
	GetRemote(resource Resource) (*Resource, error)

	// ListRemote retrieves as list of UIDs of all remote resources
	ListRemote() ([]string, error)

	// Add pushes a new resource to the endpoint
	Add(resource Resource) error

	// Update pushes an existing resource to the endpoint
	Update(existing, resource Resource) error

	// Validate gets or build the uid of corresponding resource
	Validate(resource Resource) error

	// Sort sorts resources as defined by the handler
	Sort(resources Resources) Resources

	// UsesFolders identifies whether this resource lives within a folder
	UsesFolders() bool

	// Detects whether a spec-only resource is of this kind
	Detect(map[string]any) bool
}

// ListenHandler describes a handler that has the ability to watch a single
// resource for changes, and write changes to that resource to a local file
type ListenHandler interface {
	// Listen watches a resource and update local file on changes
	Listen(UID, filename string) error
}

type ProxyEndpoint struct {
	Method  string
	Url     string
	Handler func(http.ResponseWriter, *http.Request)
}

// ProxyHandler describes a handler that can be used to edit resources live via a proxied UI
type ProxyHandler interface {
	// RegisterHandlers registers HTTP handlers for proxy events
	GetProxyEndpoints(p GrizzlyServer) []ProxyEndpoint

	// ProxyURL returns a URL path for a resource on the proxy
	ProxyURL(Resource) (string, error)
}
