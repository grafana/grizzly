package grizzly

import (
	"fmt"
	"log"
)

type Resolver struct {
	Context       map[string]Resource
	ResourceKind  ResourceKind
	ResourceKinds map[string]ResourceKind
}

func NewResolver(context []Resource, kind ResourceKind, kinds []ResourceKind) Resolver {
	resourceMap := map[string]Resource{}
	for _, resource := range context {
		key := fmt.Sprintf("%s:%s", resource.Kind(), resource.Name())
		resourceMap[key] = resource
	}

	kindsMap := map[string]ResourceKind{}
	for _, kind := range kinds {
		kindsMap[kind.Kind] = kind
	}

	return Resolver{
		Context:       resourceMap,
		ResourceKind:  kind,
		ResourceKinds: kindsMap,
	}
}

func (r *Resolver) Resolve(resource Resource) (Resources, error) {
	if r.ResourceKind.ResolvedKind == "" {
		return nil, nil
	}
	spec := resource.Spec()
	template := spec["template"].(map[string]interface{})
	templatedResource := (Resource)(template)
	newResource, err := r.ResolveResource(templatedResource)
	if err != nil {
		return nil, err
	}
	newResource["kind"] = r.ResourceKind.ResolvedKind
	return Resources{newResource}, nil
}

func (r *Resolver) ResolveResource(resource Resource) (Resource, error) {
	msi := (map[string]interface{})(resource)
	newMSI, err := r.walkMap(msi, "")
	if err != nil {
		return nil, err
	}
	return (Resource)(newMSI.(map[string]interface{})), nil
}

func (r *Resolver) walkMap(m map[string]interface{}, path string) (interface{}, error) {
	cp := make(map[string]interface{})
	for k, v := range m {
		var newPath string
		if path == "" {
			newPath = k
		} else {
			newPath = fmt.Sprintf("%s.%s", path, k)
		}
		if ref := r.getReference(newPath); ref != nil {
			resolved, err := r.resolveReference(m, k, v, *ref)
			if err != nil {
				return nil, err
			}
			properties := r.getProperties(v)
			resolvedAsMap := resolved.(map[string]interface{})
			for k, v := range properties {
				resolvedAsMap[k] = v
			}
			return resolvedAsMap, nil
		} else {
			newV, err := r.walkElement(v, newPath)
			if err != nil {
				return nil, err
			}
			cp[k] = newV
		}
	}

	return cp, nil
}

func (r *Resolver) walkElement(e interface{}, path string) (interface{}, error) {

	switch t := e.(type) {
	case map[string]interface{}:
		child, err := r.walkMap(t, path)
		if err != nil {
			return nil, err
		}
		return child, nil
	case []interface{}:
		newArray := []interface{}{}
		for _, v := range t {
			newArrayElement, err := r.walkElement(v, path)
			if err != nil {
				return nil, err
			}
			newArray = append(newArray, newArrayElement)
		}
		return newArray, nil
	default:
		return e, nil
	}
}

func (r *Resolver) getProperties(v interface{}) map[string]interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		if p, ok := t["properties"]; ok {
			return p.(map[string]interface{})
		}
	}
	return nil
}

func (r *Resolver) getReference(path string) *Reference {
	for _, ref := range r.ResourceKind.References {
		if ref.Path == path {
			return &ref
		}
	}
	return nil
}

func (r *Resolver) resolveReference(parent map[string]interface{}, key string, value interface{}, refdef Reference) (interface{}, error) {
	var msi map[string]interface{}
	switch t := value.(type) {
	case map[string]interface{}:
		msi = t
	case []interface{}:
		return nil, fmt.Errorf("Reference %s is a list not map: %v", key, t)
	default:
		return nil, fmt.Errorf("Reference %s is not a map", key)
	}
	kind := msi["kind"].(string)
	name := msi["name"].(string)
	resource, err := r.getResource(kind, name)
	if err != nil {
		return nil, err
	}
	resolver := Resolver{
		Context:       r.Context,
		ResourceKind:  r.ResourceKinds[kind],
		ResourceKinds: r.ResourceKinds,
	}
	resolved, err := resolver.ResolveResource(*resource)
	if err != nil {
		return nil, err
	}
	return resolved.Spec(), nil
}

func (r *Resolver) getResource(kind, name string) (*Resource, error) {
	key := fmt.Sprintf("%s:%s", kind, name)
	resource, ok := r.Context[key]
	if !ok {
		for k, _ := range r.Context {
			log.Printf("%s: %s", key, k)
		}
		return nil, fmt.Errorf("no resource of kind %s and name %s found", kind, name)
	}
	return &resource, nil
}
