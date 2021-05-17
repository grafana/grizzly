package grizzly

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"gopkg.in/yaml.v3"
)

// Resource represents a single Resource destined for a single endpoint
type Resource map[string]interface{}

// NewResource returns a new Resource object
func NewResource(apiVersion, kind, name string, spec map[string]interface{}) Resource {
	resource := Resource{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": spec,
	}
	return resource
}

// APIVersion returns the group and version of the provider of the resource
func (r *Resource) APIVersion() string {
	return (*r)["apiVersion"].(string)
}

// Kind returns the 'kind' of the resource, i.e. the type of the handler
func (r *Resource) Kind() string {
	return (*r)["kind"].(string)
}

func (r *Resource) Name() string {
	return r.GetMetadata("name")
}

func (r Resource) String() string {
	return r.Key()
}

// Key returns a key that combines kind and uid
func (r *Resource) Key() string {
	return fmt.Sprintf("%s/%s", r.Kind(), r.Name())
}

func (r *Resource) GetMetadata(key string) string {
	metadata := (*r)["metadata"].(map[string]interface{})
	return metadata[key].(string)
}

func (r *Resource) SetMetadata(key, value string) {
	metadata := (*r)["metadata"].(map[string]interface{})
	metadata[key] = value
	(*r)["metadata"] = metadata
}

func (r *Resource) GetSpecString(key string) string {
	spec := (*r)["spec"].(map[string]interface{})
	return spec[key].(string)
}

func (r *Resource) SetSpecString(key, value string) {
	spec := (*r)["spec"].(map[string]interface{})
	spec[key] = value
	(*r)["spec"] = spec
}

func (r *Resource) GetSpecValue(key string) interface{} {
	spec := (*r)["spec"].(map[string]interface{})
	return spec[key]
}

func (r *Resource) SetSpecValue(key string, value interface{}) {
	spec := (*r)["spec"].(map[string]interface{})
	spec[key] = value
	(*r)["spec"] = spec
}

func (r *Resource) DeleteSpecKey(key string) {
	spec := (*r)["spec"].(map[string]interface{})
	delete(spec, key)
	(*r)["spec"] = spec
}

func (r *Resource) Spec() map[string]interface{} {
	return (*r)["spec"].(map[string]interface{})
}

func (r *Resource) SpecAsJSON() (string, error) {
	j, err := json.Marshal(r.Spec())
	if err != nil {
		return "", err
	}
	return string(j), nil

}

// YAML Gets the string representation for this resource
func (r *Resource) YAML() (string, error) {
	y, err := yaml.Marshal(*r)
	if err != nil {
		return "", err
	}
	return string(y), nil
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

// Resources represents a set of resources
type Resources []Resource

// Handler describes a handler for a single API resource handled by a single provider
type Handler interface {
	APIVersion() string
	Kind() string
	GetExtension() string

	// FindResourceFiles identifies files within a directory that this handler can process
	FindResourceFiles(dir string) ([]string, error)

	// ResourceFilePath returns the location on disk where a resource should be updated
	ResourceFilePath(resource Resource, filetype string) string

	// Parse parses a manifest object into a struct for this resource type
	Parse(m manifest.Manifest) (Resources, error)

	// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
	Unprepare(resource Resource) *Resource

	// Prepare gets a resource ready for dispatch to the remote endpoint
	Prepare(existing, resource Resource) *Resource

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
	Providers []Provider
	Handlers  map[string]Handler
}

// NewProviderRegistry returns a new registry instance
func NewProviderRegistry() Registry {
	registry := Registry{}
	registry.Providers = []Provider{}
	registry.Handlers = map[string]Handler{}
	return registry
}

// RegisterProvider will register a new provider
func (r *Registry) RegisterProvider(provider Provider) error {
	r.Providers = append(r.Providers, provider)
	for _, handler := range provider.GetHandlers() {
		r.Handlers[handler.Kind()] = handler
	}
	return nil
}

// GetHandler returns a single provider based upon a JSON path
func (r *Registry) GetHandler(path string) (Handler, error) {
	handler, exists := r.Handlers[path]
	if !exists {
		return nil, fmt.Errorf("No handler registered to %s", path)
	}
	return handler, nil
}

// HandlerMatchesTarget identifies whether a handler is in a target list
func (r *Registry) HandlerMatchesTarget(handler Handler, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := handler.Kind()

	for _, target := range targets {
		if strings.Contains(target, "/") && strings.Split(target, "/")[0] == key {
			return true
		}
	}
	return false
}

// ResourceMatchesTarget identifies whether a resource is in a target list
func (r *Registry) ResourceMatchesTarget(handler Handler, UID string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := fmt.Sprintf("%s/%s", handler.Kind(), UID)
	for _, target := range targets {
		g := glob.MustCompile(target)
		if g.Match(key) {
			return true
		}
	}
	return false
}

// Notifier returns a notifier for responding to users
func (r *Registry) Notifier() *Notifier {
	return &Notifier{}
}
