package grizzly

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type ResourceRef struct {
	Kind string
	Name string
}

func NewResourceRef(kind string, name string) ResourceRef {
	return ResourceRef{
		Kind: kind,
		Name: name,
	}
}

func (ref ResourceRef) Equal(other ResourceRef) bool {
	return ref.Kind == other.Kind && ref.Name == other.Name
}

func (ref ResourceRef) String() string {
	return fmt.Sprintf("%s.%s", ref.Kind, ref.Name)
}

// Source represents the on disk (etc) location of a resource
type Source struct {
	Format     string
	Location   string
	Path       string
	Rewritable bool
}

// Resource represents a single Resource destined for a single endpoint
type Resource struct {
	Body map[string]interface{}

	Source Source
}

func ResourceFromMap(data map[string]interface{}) (*Resource, error) {
	r := Resource{
		Body: data,
	}

	// Ensure that the spec is a map
	spec := r.Body["spec"]
	if spec == nil {
		return nil, fmt.Errorf("resource %s has no spec", r.Name())
	}
	if _, isMap := spec.(map[string]interface{}); !isMap {
		return nil, fmt.Errorf("resource %s has an invalid spec. Expected a map, got a value of type %T", r.Name(), spec)
	}

	return &r, nil
}

// NewResource returns a new Resource object
func NewResource(apiVersion, kind, name string, spec map[string]interface{}) (Resource, error) {
	resource := Resource{
		Body: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": spec,
		},
	}
	return resource, nil
}

// APIVersion returns the group and version of the provider of the resource
func (r *Resource) APIVersion() string {
	return r.Body["apiVersion"].(string)
}

// Kind returns the 'kind' of the resource, i.e. the type of the handler
func (r *Resource) Kind() string {
	return r.Body["kind"].(string)
}

func (r *Resource) Name() string {
	return r.GetMetadata("name")
}

func (r *Resource) Ref() ResourceRef {
	return ResourceRef{
		Kind: r.Kind(),
		Name: r.Name(),
	}
}

func (r *Resource) SetSource(source Source) {
	r.Source = source
}

func (r Resource) String() string {
	return r.Ref().String()
}

func (r Resource) metadata() map[string]any {
	return r.Body["metadata"].(map[string]interface{})
}

func (r *Resource) HasMetadata(key string) bool {
	_, ok := r.metadata()[key]
	return ok
}

func (r *Resource) GetMetadata(key string) string {
	if !r.HasMetadata(key) {
		return ""
	}
	return r.metadata()[key].(string)
}

func (r *Resource) SetMetadata(key, value string) {
	metadata := r.metadata()
	metadata[key] = value
	r.Body["metadata"] = metadata
}

func (r *Resource) GetSpecString(key string) (string, bool) {
	if val, ok := r.Spec()[key]; ok {
		return val.(string), true
	}
	return "", false
}

func (r *Resource) SetSpecString(key, value string) {
	spec := r.Spec()
	spec[key] = value
	r.Body["spec"] = spec
}

func (r *Resource) GetSpecValue(key string) interface{} {
	return r.Spec()[key]
}

func (r *Resource) SetSpecValue(key string, value interface{}) {
	spec := r.Spec()
	spec[key] = value
	r.Body["spec"] = spec
}

func (r *Resource) DeleteSpecKey(key string) {
	spec := r.Spec()
	delete(spec, key)
	r.Body["spec"] = spec
}

func (r *Resource) Spec() map[string]interface{} {
	return r.Body["spec"].(map[string]interface{})
}

func (r *Resource) SpecAsJSON() (string, error) {
	j, err := json.MarshalIndent(r.Spec(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil

}

// YAML Gets the string representation for this resource
func (r *Resource) YAML() (string, error) {
	y, err := yaml.Marshal(r.Spec())
	if err != nil {
		return "", err
	}
	return string(y), nil
}

// Resources represents a set of resources
type Resources []Resource

func (r Resources) Find(ref ResourceRef) (Resource, bool) {
	for _, resource := range r {
		if resource.Ref().Equal(ref) {
			return resource, true
		}
	}

	return Resource{}, false
}

func (r Resources) Filter(predicate func(Resource) bool) Resources {
	filtered := make(Resources, 0)

	for _, resource := range r {
		if predicate(resource) {
			filtered = append(filtered, resource)
		}
	}

	return filtered
}

func (r Resources) Len() int {
	return len(r)
}
