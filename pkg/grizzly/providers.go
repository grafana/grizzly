package grizzly

import (
	"encoding/json"
	"fmt"

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
	uid := r.UID()
	return fmt.Sprintf("%s.%s", r.Kind(), uid)
}

func (r Resource) UID() string {
	handler, err := Registry.GetHandler(r.Kind())
	if err != nil {
		return "Unknown-handler:" + r.Kind()
	}
	uid, err := handler.GetUID(r)
	if err != nil {
		return "error:" + err.Error()
	}
	return uid
}

func (r *Resource) HasMetadata(key string) bool {
	metadata := (*r)["metadata"].(map[string]interface{})
	_, ok := metadata[key]
	return ok
}

func (r *Resource) GetMetadata(key string) string {
	metadata := (*r)["metadata"].(map[string]interface{})
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return value
}

func (r *Resource) SetMetadata(key, value string) {
	metadata := (*r)["metadata"].(map[string]interface{})
	metadata[key] = value
	(*r)["metadata"] = metadata
}

func (r *Resource) GetSpecString(key string) (string, bool) {
	spec := (*r)["spec"].(map[string]interface{})
	if val, ok := spec[key]; ok {
		return val.(string), true
	}
	return "", false
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
	UID := r.UID()
	dotKey := r.Key()
	slashKey := fmt.Sprintf("%s/%s", r.Kind(), UID)
	for _, target := range targets {
		g := glob.MustCompile(target)
		if g.Match(slashKey) || g.Match(dotKey) {
			return true
		}
	}
	return false
}

// Resources represents a set of resources
type Resources []Resource

func (r Resources) Len() int {
	return len(r)
}

func (r Resources) Less(i, j int) bool {
	iKind := r[i].Kind()
	jKind := r[j].Kind()
	iPos := Registry.HandlerOrder[iKind]
	jPos := Registry.HandlerOrder[jKind]
	return iPos < jPos
}

func (r Resources) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

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

	// Retrieves a UID for a resource
	GetUID(resource Resource) (string, error)

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
}

// PreviewHandler describes a handler that has the ability to render
// a preview of a resource
type PreviewHandler interface {
	// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
	Preview(resource Resource, opts *PreviewOpts) error
}

// ListenHandler describes a handler that has the ability to watch a single
// resource for changes, and write changes to that resource to a local file
type ListenHandler interface {
	// Listen watches a resource and update local file on changes
	Listen(UID, filename string) error
}
