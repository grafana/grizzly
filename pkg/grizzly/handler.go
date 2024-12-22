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

func (h *BaseHandler) Prepare(existing *Resource, resource Resource) *Resource {
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

	// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
	Unprepare(resource Resource) *Resource

	// Prepare gets a resource ready for dispatch to the remote endpoint
	Prepare(existing *Resource, resource Resource) *Resource

	// GetUID retrieves a UID for a resource
	GetUID(resource Resource) (string, error)

	// GetSpecUID retrieves a UID from the spec of a raw resource
	GetSpecUID(resource Resource) (string, error)

	// GetByUID retrieves JSON for a resource from an endpoint, by UID
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

	// Detect whether a spec-only resource is of this kind
	Detect(map[string]any) bool
}

// SnapshotHandler describes a handler that has the ability to push a resource as
// a snapshot
type SnapshotHandler interface {
	// Snapshot pushes a resource as a snapshot with an expiry
	Snapshot(resource Resource, expiresSeconds int) error
}

// ListenHandler describes a handler that has the ability to watch a single
// resource for changes, and write changes to that resource to a local file
type ListenHandler interface {
	// Listen watches a resource and update local file on changes
	Listen(UID, filename string) error
}

// ProxyConfiguratorProvider indicates that the handler implementing it
// provides configuration on how to proxy the resources it manages.
type ProxyConfiguratorProvider interface {
	ProxyConfigurator() ProxyConfigurator
}

type HTTPEndpoint struct {
	Method  string
	URL     string
	Handler http.HandlerFunc
}

// StaticProxyConfig holds some static configuration to apply to the proxy.
// This allows resource handlers to declare routes to proxy or mock that are
// specific to them.
type StaticProxyConfig struct {
	// ProxyGet holds a list of routes to proxy when using the GET HTTP
	// method.
	// Example: /public/*
	ProxyGet []string

	// ProxyPost holds a list of routes to proxy when using the POST HTTP
	// method.
	// Example: /api/v1/eval
	ProxyPost []string

	// MockGet holds a map associating URLs to a mock response that they should
	// return for GET requests.
	// Note: the response is expected to be JSON.
	MockGet map[string]string

	// MockPost holds a map associating URLs to a mock response that they should
	// return for POST requests.
	// Note: the response is expected to be JSON.
	MockPost map[string]string
}

// ProxyConfigurator describes a proxy endpoints that can be used to view/edit
// resources live via a proxied UI.
type ProxyConfigurator interface {
	// Endpoints registers HTTP handlers for proxy events
	Endpoints(p Server) []HTTPEndpoint

	// ProxyURL returns a URL path for a resource on the proxy
	ProxyURL(uid string) string

	// ProxyEditURL returns a URL path for a resource on the proxy
	ProxyEditURL(uid string) string

	StaticEndpoints() StaticProxyConfig
}
