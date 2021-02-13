package grizzly

import "fmt"

// Resource represents a single Resource destined for a single endpoint
type Resource struct {
	UID      string      `json:"uid"`
	Filename string      `json:"filename"`
	Handler  Handler     `json:"handler"`
	Detail   interface{} `json:"detail"`
	JSONPath string      `json:"path"`
}

// Kind returns the 'kind' of the resource, i.e. the type of the handler
func (r *Resource) Kind() string {
	return r.Handler.GetFullName()
}

// Key returns a key that combines kind and uid
func (r *Resource) Key() string {
	return fmt.Sprintf("%s/%s", r.Kind(), r.UID)
}

// GetRepresentation Gets the string representation for this resource
func (r *Resource) GetRepresentation() (string, error) {
	return r.Handler.GetRepresentation(r.UID, *r)
}

// GetRemoteRepresentation Gets the string representation for this resource
func (r *Resource) GetRemoteRepresentation() (string, error) {
	return r.Handler.GetRemoteRepresentation(r.UID)
}

// MatchesTarget identifies whether a resource is in a target list
func (r *Resource) MatchesTarget(targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := r.Key()
	for _, target := range targets {
		if target == key {
			return true
		}
	}
	return false
}

// ResourceList represents a set of named resources
type ResourceList map[string]Resource

// Resources represents a set of resources by handler
type Resources map[Handler]ResourceList

// Handler describes a handler for a single API resource handled by a single provider
type Handler interface {
	GetName() string
	GetFullName() string
	GetProvider() string
	GetJSONPaths() []string
	GetExtension() string

	// Parse parses an interface{} object into a struct for this resource type
	Parse(path string, i interface{}) (ResourceList, error)

	// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
	Unprepare(resource Resource) *Resource

	// Prepare gets a resource ready for dispatch to the remote endpoint
	Prepare(existing, resource Resource) *Resource

	// Get retrieves JSON for a resource from an endpoint, by UID
	GetByUID(UID string) (*Resource, error)

	// GetRepresentation renders Jsonnet to Grizzly resources, rendering as a string
	GetRepresentation(uid string, resource Resource) (string, error)

	// GetRemoteRepresentation retrieves a resource from the endpoint and renders to a string
	GetRemoteRepresentation(uid string) (string, error)

	// GetRemote retrieves a resource as a datastructure
	GetRemote(uid string) (*Resource, error)

	// Add pushes a new resource to the endpoint
	Add(resource Resource) error

	// Update pushes an existing resource to the endpoint
	Update(existing, resource Resource) error
}

// MultiResourceHandler describes a handler that can handle multiple resources in one go.
// This could be because it needs to see all resources before sending, or because the
// endpoint API supports batching of resources.
type MultiResourceHandler interface {
	// Diff compares local resources with remote equivalents and output result
	Diff(notifier Notifier, resources ResourceList) error

	// Apply local resources to remote endpoint
	Apply(notifier Notifier, resources ResourceList) error
}

// PreviewHandler describes a handler that has the ability to render
// a preview of a resource
type PreviewHandler interface {
	// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
	Preview(resource Resource, notifier Notifier, opts *PreviewOpts) error
}

// ListenHandler describes a handler that has the ability to watch a single
// resource for changes, and write changes to that resource to a local file
type ListenHandler interface {
	// Listen watches a resource and update local file on changes
	Listen(notifier Notifier, UID, filename string) error
}

// Provider describes a single Endpoint Provider
type Provider interface {
	GetName() string
	GetHandlers() []Handler
}

// Registry records providers
type Registry struct {
	Providers     []Provider
	Handlers      []Handler
	HandlerByName map[string]Handler
	HandlerByPath map[string]Handler
}

// NewProviderRegistry returns a new registry instance
func NewProviderRegistry() Registry {
	registry := Registry{}
	registry.Providers = []Provider{}
	registry.Handlers = []Handler{}
	registry.HandlerByName = map[string]Handler{}
	registry.HandlerByPath = map[string]Handler{}
	return registry
}

// RegisterProvider will register a new provider
func (r *Registry) RegisterProvider(provider Provider) error {
	r.Providers = append(r.Providers, provider)
	for _, handler := range provider.GetHandlers() {
		r.Handlers = append(r.Handlers, handler)
		for _, path := range handler.GetJSONPaths() {
			r.HandlerByPath[path] = handler
		}
		r.HandlerByName[handler.GetName()] = handler
		r.HandlerByName[handler.GetFullName()] = handler
	}
	return nil
}

// GetHandler returns a single provider based upon a JSON path
func (r *Registry) GetHandler(path string) (Handler, error) {
	handler, exists := r.HandlerByPath[path]
	if !exists {
		handler, exists = r.HandlerByName[path]
		if !exists {
			return nil, fmt.Errorf("No handler registered to %s", path)
		}
		return handler, nil
	}
	return handler, nil
}
