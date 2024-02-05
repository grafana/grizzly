package grizzly

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Resource represents a single Resource destined for a single endpoint
type Resource map[string]interface{}

func ResourceFromMap(data map[string]interface{}) (Resource, error) {
	r := Resource(data)

	// Ensure that the spec is a map
	spec := r["spec"]
	if spec == nil {
		return nil, fmt.Errorf("resource %s has no spec", r.Name())
	}
	if _, isMap := spec.(map[string]interface{}); !isMap {
		return nil, fmt.Errorf("resource %s has an invalid spec. Expected a map, got a value of type %T", r.Name(), spec)
	}

	return r, nil
}

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
	return metadata[key].(string)
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
	j, err := json.MarshalIndent(r.Spec(), "", "  ")
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

// JSON Gets the string representation for this resource
func (r *Resource) JSON() (string, error) {
	j, err := json.MarshalIndent(*r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// Resources represents a set of resources
type Resources []Resource

func (r Resources) Len() int {
	return len(r)
}

func (r *Resources) Sort() Resources {
	resourceByKind := map[string]Resources{}
	resources := Resources{}
	for _, resource := range *r {
		resourceByKind[resource.Kind()] = append(resourceByKind[resource.Kind()], resource)
	}
	for _, handler := range Registry.HandlerOrder {
		handlerResources := resourceByKind[handler.Kind()]
		resources = append(resources, handler.Sort(handlerResources)...)
	}
	return resources
}
