package grizzly

import (
	"fmt"

	"github.com/gobwas/glob"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"gopkg.in/yaml.v2"
)

// Resource represents a single Resource destined for a single endpoint
type Resource struct {
	UID      string            `json:"uid"`
	Filename string            `json:"filename"`
	Handler  Handler           `json:"handler"`
	Detail   manifest.Manifest `json:"detail"`
}

// NewResource encapsulates a manifest into a Resource object
func NewResource(m manifest.Manifest, handler Handler) *Resource {
	resource := Resource{
		UID:     m.Metadata().Name(),
		Detail:  m,
		Handler: handler,
	}
	return &resource
}

// APIVersion returns the group and version of the provider of the resource
func (r *Resource) APIVersion() string {
	return r.Handler.APIVersion()
}

// Kind returns the 'kind' of the resource, i.e. the type of the handler
func (r *Resource) Kind() string {
	return r.Handler.Kind()
}

// Key returns a key that combines kind and uid
func (r *Resource) Key() string {
	return fmt.Sprintf("%s/%s", r.Kind(), r.UID)
}

// GetRepresentation Gets the string representation for this resource
func (r *Resource) GetRepresentation() (string, error) {
	y, err := yaml.Marshal(r.Detail)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

// GetRemoteRepresentation Gets the string representation for this resource
func (r *Resource) GetRemoteRepresentation() (string, error) {
	remote, err := r.Handler.GetRemote(*r)
	if err != nil {
		return "", err
	}
	return remote.GetRepresentation()
}

// MatchesTarget identifies whether a resource is in a target list
func (r *Resource) MatchesTarget(targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := r.Key()
	for _, target := range targets {
		g := glob.MustCompile(target)
		if g.Match(key) {
			return true
		}
	}
	return false
}

// DeleteSpecKey deletes an element from the spec
func (r *Resource) DeleteSpecKey(key string) {
	msi := r.Detail["spec"].(map[string]interface{})
	delete(msi, key)
	r.Detail["spec"] = msi
}

// GetSpecKey retrieves a value from the spec
func (r *Resource) GetSpecKey(key string) string {
	msi := r.Detail["spec"].(map[string]interface{})
	return msi[key].(string)
}

// SetSpecKey sets a value from in the spec
func (r *Resource) SetSpecKey(key, value string) {
	msi := r.Detail["spec"].(map[string]interface{})
	msi[key] = value
	r.Detail["spec"] = msi
}

// ResourceList represents a set of named resources
type ResourceList map[string]Resource

// Resources represents a set of resources by handler
type Resources map[Handler]ResourceList

// Handler describes a handler for a single API resource handled by a single provider
type Handler interface {
	APIVersion() string
	Kind() string
	GetExtension() string

	// GetRemoteByUID retrieves a remote resource identified by a string
	GetRemoteByUID(uid string) (*Resource, error)

	// GetRemote retrieves a remote equivalent of a resource as a datastructure
	GetRemote(resource Resource) (*Resource, error)

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
	Group() string
	Version() string
	APIVersion() string
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
		r.HandlerByName[handler.Kind()] = handler
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
