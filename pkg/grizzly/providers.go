package grizzly

import "fmt"

// Resource represents a single Resource destined for a single endpoint
type Resource struct {
	UID      string                 `json:"uid"`
	Filename string                 `json:"filename"`
	Provider Provider               `json:"provider"`
	Detail   map[string]interface{} `json:"detail"`
	Path     string                 `json:"path"`
}

// Kind returns the 'kind' of the resource, i.e. the type of the provider
func (r *Resource) Kind() string {
	return r.Provider.GetName()
}

// Key returns a key that combines kind and uid
func (r *Resource) Key() string {
	return fmt.Sprintf("%s/%s", r.Kind(), r.UID)
}

// GetRepresentation Gets the string representation for this resource
func (r *Resource) GetRepresentation() (string, error) {
	return r.Provider.GetRepresentation(r.UID, r.Detail)
}

// GetRemoteRepresentation Gets the string representation for this resource
func (r *Resource) GetRemoteRepresentation() (string, error) {
	return r.Provider.GetRemoteRepresentation(r.UID)
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

// Resources represents a set of resources, by path
type Resources map[string]Resource

// Provider describes a single Endpoint Provider
type Provider interface {
	GetName() string
	GetJSONPath() string
	GetExtension() string

	// Parse parses an interface{} object into a struct for this resource type
	Parse(i interface{}) (Resources, error)

	// Get retrieves JSON for a resource from an endpoint, by UID
	GetByUID(UID string) (*Resource, error)

	// GetRepresentation renders Jsonnet to Grizzly resources, rendering as a string
	GetRepresentation(uid string, detail map[string]interface{}) (string, error)

	// GetRemoteRepresentation retrieves a resource from the endpoint and renders to a string
	GetRemoteRepresentation(uid string) (string, error)

	// GetRemote retrieves a resource as a datastructure
	GetRemote(uid string) (*Resource, error)

	// Add pushes a new resource to the endpoint
	Add(detail map[string]interface{}) error

	// Update pushes an existing resource to the endpoint
	Update(current, detail map[string]interface{}) error

	// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
	Preview(detail map[string]interface{}, opts *PreviewOpts) error
}

// Registry records providers
type Registry struct {
	ProviderMap  map[string]Provider
	ProviderList []Provider
}

// NewProviderRegistry returns a new registry instance
func NewProviderRegistry() Registry {
	registry := Registry{}
	registry.ProviderMap = map[string]Provider{}
	registry.ProviderList = []Provider{}
	return registry
}

// RegisterProvider will register a new provider
func (r *Registry) RegisterProvider(provider Provider) error {
	path := provider.GetJSONPath()
	r.ProviderMap[path] = provider
	r.ProviderList = append(r.ProviderList, provider)
	return nil
}

// GetProviders will retrieve all registered providers
func (r *Registry) GetProviders() []Provider {
	return r.ProviderList
}

// GetProvider returns a single provider based upon a JSON path
func (r *Registry) GetProvider(jsonPath string) (Provider, error) {
	provider, exists := r.ProviderMap[jsonPath]
	if !exists {
		return nil, fmt.Errorf("No provider registered to path %s", jsonPath)
	}
	return provider, nil
}
